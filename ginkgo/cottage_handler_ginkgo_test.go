package ginkgo_test

import (
	"bytes"
	"context"
	"cottageManager/internal/app"
	"cottageManager/internal/config"
	"cottageManager/internal/domain"
	mdb "cottageManager/internal/infra/db"
	"cottageManager/internal/infra/logging"
	"cottageManager/internal/transport/http/availability"
	"cottageManager/internal/transport/http/booking"
	"cottageManager/internal/transport/http/cottage"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	testDatabaseName = "ginkgo_cottage_test"
)

var (
	mongoContainer *mongodb.MongoDBContainer
	mongoClient    *mongo.Client
	db             *mongo.Database
	router         *gin.Engine
	logShutdown    func(context.Context) error
)

var _ = BeforeSuite(func() {
	gin.SetMode(gin.TestMode)
	var err error
	logShutdown, err = logging.Setup(context.Background(), &config.LogConfig{
		Level:  "info",
		Format: "text",
	}, nil)
	Expect(err).NotTo(HaveOccurred())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mongoContainer, err = mongodb.Run(
		ctx,
		"mongo:latest",
		mongodb.WithUsername("test_user"),
		mongodb.WithPassword("test_pass"),
	)
	Expect(err).NotTo(HaveOccurred())

	uri, err := mongoContainer.ConnectionString(ctx)
	Expect(err).NotTo(HaveOccurred())

	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	Expect(err).NotTo(HaveOccurred())

	db = mongoClient.Database(testDatabaseName)

	cottageRepo := mdb.NewCottageRepo(db, &config.CottageCollectionConfig{Name: "cottage"})
	bookingRepo := mdb.NewBookingRepo(db, &config.BookingCollectionConfig{Name: "booking"})

	availabilityService := app.NewAvailabilityService(cottageRepo, bookingRepo)
	cottageService := app.NewCottageService(cottageRepo)
	bookingService := app.NewBookingService(availabilityService, cottageService, bookingRepo)

	availabilityHandler := availability.NewHandler(availabilityService)
	bookingHandler := booking.NewHandler(bookingService)
	cottageHandler := cottage.NewHandler(cottageService)

	router = gin.New()

	// return the list with details of all cottages
	router.GET("/cottages", cottageHandler.GetAll)
	// return the cottage for the given name
	router.GET("/cottage/:name", cottageHandler.GetByName)

	// return a list of available periods in the given date range
	router.GET("/cottage/:name/available-dates", availabilityHandler.GetAvailablePeriods)
	// return a list of available periods for every cottage of a given type in the given date range
	router.GET("/cottage/type/:cottageType/available-dates", availabilityHandler.GetAvailablePeriodsByCottageType)

	// register a booking for a cottage
	router.POST("/cottage/:name/booking", bookingHandler.AddBooking)
	// delete a booking for a given cottage
	router.DELETE("/cottage/:name/booking/:bookingId", bookingHandler.RemoveBooking)

})

var _ = AfterSuite(func() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if mongoClient != nil {
		_ = mongoClient.Disconnect(ctx)
	}

	if mongoContainer != nil {
		_ = testcontainers.TerminateContainer(mongoContainer)
	}

	if logShutdown != nil {
		_ = logShutdown(context.Background())
	}
})

var _ = Describe("Scenario: client tries to book a cottage", Ordered, func() {
	BeforeAll(func() {
		collection := db.Collection("cottage")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		db = mongoClient.Database(testDatabaseName)
		_ = db.Drop(ctx)

		data, err := os.ReadFile("test_data/cottages.json")
		if err != nil {
			Expect(err).NotTo(HaveOccurred())
		}

		var items []domain.Cottage
		if err := json.Unmarshal(data, &items); err != nil {
			Expect(err).NotTo(HaveOccurred())
		}

		docs := make([]interface{}, len(items))
		for i, g := range items {
			docs[i] = g
		}
		if _, err := collection.InsertMany(context.Background(), docs); err != nil {
			Expect(err).NotTo(HaveOccurred())
		}
	})

	AfterAll(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if db != nil {
			_ = db.Drop(ctx)
		}
	})

	var cottages []cottage.Dto
	var lilyOfTheValleyCottage cottage.Dto
	var lilyAvailablePeriods []availability.AvailablePeriodDTO
	var bookingConfirmation booking.ConfirmationDto

	It("client browsers all cottages", func() {
		req := httptest.NewRequest(http.MethodGet, "/cottages", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(http.StatusOK))

		if err := json.Unmarshal(w.Body.Bytes(), &cottages); err != nil {
			Expect(err).NotTo(HaveOccurred())
		}

		cottageNames := make([]string, len(cottages))
		for i, c := range cottages {
			cottageNames[i] = c.Name
		}
		Expect(cottageNames).To(ContainElements("Barbara Karst", "Juanita Hatten", "Golden Glow", "Singapore White",
			"Raspberry Ice", "Torch Ginger", "Bird of Paradise", "Passion Flower", "Wild Rose", "Lily of the Valley"))
	})

	It("client choose to see Lily of the Valley cottage", func() {
		req := httptest.NewRequest(http.MethodGet, "/cottage/Lily%20of%20the%20Valley", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if err := json.Unmarshal(w.Body.Bytes(), &lilyOfTheValleyCottage); err != nil {
			Expect(err).NotTo(HaveOccurred())
		}

		Expect(w.Code).To(Equal(http.StatusOK))
		Expect(lilyOfTheValleyCottage.Name).To(Equal("Lily of the Valley"))
	})

	It("client checks the availability of Lily of the Valley cottage for next week", func() {
		nextWeek := time.Now().UTC().AddDate(0, 0, 7)
		weekAfterNextWeek := nextWeek.AddDate(0, 0, 7)

		periodStart := nextWeek.Format("2006-01-02")
		periodEnd := weekAfterNextWeek.Format("2006-01-02")

		req := httptest.NewRequest(http.MethodGet, "/cottage/Lily%20of%20the%20Valley/available-dates?from="+periodStart+"&to="+periodEnd, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if err := json.Unmarshal(w.Body.Bytes(), &lilyAvailablePeriods); err != nil {
			Expect(err).NotTo(HaveOccurred())
		}

		Expect(w.Code).To(Equal(http.StatusOK))
		Expect(len(lilyAvailablePeriods)).To(Equal(1))
	})

	It("client tries to book Lily of the Valley cottage for next week", func() {
		bookingRequest, err := json.Marshal(booking.RequestDto{
			GuestId:        primitive.NewObjectID().Hex(),
			NumberOfGuests: 2,
			CheckInDate:    lilyAvailablePeriods[0].From.Format("2006-01-02"),
			CheckOutDate:   lilyAvailablePeriods[0].To.Format("2006-01-02"),
		})

		if err != nil {
			Expect(err).NotTo(HaveOccurred())
		}

		req := httptest.NewRequest(http.MethodPost, "/cottage/Lily%20of%20the%20Valley/booking", bytes.NewReader(bookingRequest))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if err := json.Unmarshal(w.Body.Bytes(), &bookingConfirmation); err != nil {
			Expect(err).NotTo(HaveOccurred())
		}

		Expect(w.Code).To(Equal(http.StatusOK))
	})
})
