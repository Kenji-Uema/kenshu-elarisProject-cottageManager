package ginkgo_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/app"
	"github.com/Kenji-Uema/cottageManager/internal/config"
	"github.com/Kenji-Uema/cottageManager/internal/domain/document"
	"github.com/Kenji-Uema/cottageManager/internal/domain/dto"
	mdb "github.com/Kenji-Uema/cottageManager/internal/infra/db"
	"github.com/Kenji-Uema/cottageManager/internal/infra/logging"
	transport "github.com/Kenji-Uema/cottageManager/internal/transport/http"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	testDatabaseName = "ginkgo_cottage_test"
)

var (
	mongoContainer *mongodb.MongoDBContainer
	mongoClient    *mongo.Client
	db             *mongo.Database
	router         *gin.Engine
)

type noopTxManager struct{}

func (noopTxManager) WithTransaction(ctx context.Context, callback func(ctx context.Context) (any, error)) (any, error) {
	return callback(ctx)
}

var _ = BeforeSuite(func() {
	gin.SetMode(gin.TestMode)
	slog.SetDefault(logging.NewLogger())
	var err error

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

	mongoClient, err = mongo.Connect(options.Client().ApplyURI(uri))
	Expect(err).NotTo(HaveOccurred())

	db = mongoClient.Database(testDatabaseName)

	cottageRepo := mdb.NewCottageRepo(db, config.CottageCollectionConfig{Name: "cottage"})
	bookingRepo := mdb.NewBookingRepo(db, config.BookingCollectionConfig{Name: "booking"})
	txManager := noopTxManager{}

	cottageService := app.NewCottageService(cottageRepo)
	bookingService := app.NewBookingService(cottageService, bookingRepo, txManager)
	availabilityService := app.NewAvailabilityService(cottageService, bookingService)

	httpServer := transport.NewHttpServer(config.ServerConfig{})
	httpServer.SetupRoutes(
		availabilityService,
		bookingService,
		cottageService,
		&mdb.Db{Client: mongoClient, Database: db},
	)
	router = httpServer.Router

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
})

var _ = Describe("Scenario: client tries to book a cottage", Ordered, func() {
	BeforeAll(func() {
		collection := db.Collection("cottage")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		db = mongoClient.Database(testDatabaseName)
		_ = db.Drop(ctx)

		dataPath := projectRootTestDataPath()
		data, err := os.ReadFile(dataPath)
		if err != nil {
			Expect(err).NotTo(HaveOccurred())
		}

		var items []document.Cottage
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

	var cottages []dto.Cottage
	var selectedCottage dto.Cottage
	var selectedCottageAvailablePeriods dto.AvailablePeriodDTO
	var bookingConfirmation dto.ConfirmationDto

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
		Expect(cottageNames).To(ContainElements("Rose", "Lily", "Daisy"))
	})

	It("client choose to see Daisy cottage", func() {
		req := httptest.NewRequest(http.MethodGet, "/cottage/Daisy", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if err := json.Unmarshal(w.Body.Bytes(), &selectedCottage); err != nil {
			Expect(err).NotTo(HaveOccurred())
		}

		Expect(w.Code).To(Equal(http.StatusOK))
		Expect(selectedCottage.Name).To(Equal("Daisy"))
	})

	It("client checks the availability of Daisy cottage for next week", func() {
		nextWeek := time.Now().UTC().AddDate(0, 0, 7)
		weekAfterNextWeek := nextWeek.AddDate(0, 0, 7)

		periodStart := nextWeek.Format("2006-01-02")
		periodEnd := weekAfterNextWeek.Format("2006-01-02")

		req := httptest.NewRequest(http.MethodGet, "/cottage/Daisy/available-dates?from="+periodStart+"&to="+periodEnd, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if err := json.Unmarshal(w.Body.Bytes(), &selectedCottageAvailablePeriods); err != nil {
			Expect(err).NotTo(HaveOccurred())
		}

		Expect(w.Code).To(Equal(http.StatusOK))
		Expect(len(selectedCottageAvailablePeriods.Periods)).To(Equal(1))
	})

	It("client tries to book Daisy cottage for next week", func() {
		bookingRequest, err := json.Marshal(dto.BookingRequestDto{
			GuestId:        bson.NewObjectID().Hex(),
			NumberOfGuests: 2,
			CheckInDate:    selectedCottageAvailablePeriods.Periods[0].CheckIn.Format("2006-01-02"),
			CheckOutDate:   selectedCottageAvailablePeriods.Periods[0].CheckOut.Format("2006-01-02"),
		})

		if err != nil {
			Expect(err).NotTo(HaveOccurred())
		}

		req := httptest.NewRequest(http.MethodPost, "/cottage/Daisy/booking", bytes.NewReader(bookingRequest))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if err := json.Unmarshal(w.Body.Bytes(), &bookingConfirmation); err != nil {
			Expect(err).NotTo(HaveOccurred())
		}

		Expect(w.Code).To(Equal(http.StatusOK))
	})
})

func projectRootTestDataPath() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "../test_data/cottages.json"
	}

	// this file is in /ginkgo; project root is one level up
	projectRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), ".."))
	return filepath.Join(projectRoot, "test_data", "cottages.json")
}
