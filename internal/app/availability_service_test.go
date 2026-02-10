package app

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/Kenji-Uema/cottageManager/internal/domain"
	portmocks "github.com/Kenji-Uema/cottageManager/internal/port/mocks"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var initAvailabilityServiceMocks = func() (*portmocks.MockCottageRepo, *portmocks.MockBookingRepo) {
	cr := portmocks.NewMockCottageRepo()
	br := portmocks.NewMockBookingRepo()

	return cr, br
}

func Test_availabilityService_GetAvailablePeriods(t *testing.T) {
	requestPeriod := domain.Period{
		Start: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC),
	}

	t.Run("when cottage X has no bookings, then the whole period should be returned", func(t *testing.T) {
		cr, br := initAvailabilityServiceMocks()

		cr.GetByNameFunc = func(_ context.Context, _ string) (domain.Cottage, error) {
			return domain.Cottage{}, nil
		}
		br.GetBookingsFunc = func(_ context.Context, _ []bson.ObjectID) ([]domain.Booking, error) {
			return []domain.Booking{}, nil
		}

		wanted := []domain.Period{
			{
				Start: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
			},
		}

		s := NewAvailabilityService(cr, br)
		got, _ := s.GetAvailablePeriods(context.Background(), "X", requestPeriod)

		if !reflect.DeepEqual(got, wanted) {
			t.Errorf("GetAvailablePeriods() got = %v, want %v", got, wanted)
		}

	})

	t.Run("when cottage X has 1 booking covering the whole period, then should return 0 available period", func(t *testing.T) {
		cr, br := initAvailabilityServiceMocks()

		booking := domain.Booking{StayPeriod: domain.Period{
			Start: time.Date(2025, 8, 5, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 10, 10, 0, 0, 0, 0, time.UTC),
		}}

		cr.GetByNameFunc = func(_ context.Context, _ string) (domain.Cottage, error) {
			return domain.Cottage{}, nil
		}
		br.GetBookingsFunc = func(_ context.Context, _ []bson.ObjectID) ([]domain.Booking, error) {
			return []domain.Booking{booking}, nil
		}

		s := NewAvailabilityService(cr, br)
		got, _ := s.GetAvailablePeriods(context.Background(), "X", requestPeriod)

		if !reflect.DeepEqual(got, []domain.Period{}) {
			t.Errorf("GetAvailablePeriods() got = %v, want %v", got, []domain.Period{})
		}
	})

	t.Run("when cottage X has 2 bookings covering the whole period, then should return 0 available period", func(t *testing.T) {
		cr, br := initAvailabilityServiceMocks()

		booking1 := domain.Booking{StayPeriod: domain.Period{
			Start: time.Date(2025, 8, 30, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC),
		}}
		booking2 := domain.Booking{StayPeriod: domain.Period{
			Start: time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 10, 10, 0, 0, 0, 0, time.UTC),
		}}

		cr.GetByNameFunc = func(_ context.Context, _ string) (domain.Cottage, error) {
			return domain.Cottage{}, nil
		}
		br.GetBookingsFunc = func(_ context.Context, _ []bson.ObjectID) ([]domain.Booking, error) {
			return []domain.Booking{booking1, booking2}, nil
		}

		s := NewAvailabilityService(cr, br)
		got, _ := s.GetAvailablePeriods(context.Background(), "X", requestPeriod)

		if !reflect.DeepEqual(got, []domain.Period{}) {
			t.Errorf("GetAvailablePeriods() got = %v, want %v", got, []domain.Period{})
		}
	})

	t.Run("when cottage X has 1 booking inside the requested period, then should return 2 available periods", func(t *testing.T) {
		cr, br := initAvailabilityServiceMocks()

		booking := domain.Booking{StayPeriod: domain.Period{
			Start: time.Date(2025, 9, 5, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC),
		}}

		cr.GetByNameFunc = func(_ context.Context, _ string) (domain.Cottage, error) {
			return domain.Cottage{}, nil
		}
		br.GetBookingsFunc = func(_ context.Context, _ []bson.ObjectID) ([]domain.Booking, error) {
			return []domain.Booking{booking}, nil
		}

		s := NewAvailabilityService(cr, br)
		got, _ := s.GetAvailablePeriods(context.Background(), "X", requestPeriod)

		wanted := []domain.Period{
			{
				Start: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 9, 5, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
			},
			{
				Start: time.Date(2025, 9, 11, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
			},
		}

		if !reflect.DeepEqual(got, wanted) {
			t.Errorf("GetAvailablePeriods() got = %v, want %v", got, wanted)
		}
	})

	t.Run("when cottage X has 1 booking overlapping the beginning of the period, then should return 1 available period", func(t *testing.T) {
		cr, br := initAvailabilityServiceMocks()

		booking := domain.Booking{StayPeriod: domain.Period{
			Start: time.Date(2025, 8, 5, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC),
		}}

		cr.GetByNameFunc = func(_ context.Context, _ string) (domain.Cottage, error) {
			return domain.Cottage{}, nil
		}
		br.GetBookingsFunc = func(_ context.Context, _ []bson.ObjectID) ([]domain.Booking, error) {
			return []domain.Booking{booking}, nil
		}

		s := NewAvailabilityService(cr, br)
		got, _ := s.GetAvailablePeriods(context.Background(), "X", requestPeriod)

		wanted := []domain.Period{
			{
				Start: time.Date(2025, 9, 11, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
			},
		}

		if !reflect.DeepEqual(got, wanted) {
			t.Errorf("GetAvailablePeriods() got = %v, want %v", got, wanted)
		}
	})

	t.Run("when cottage X has 1 booking overlapping the end of the period, then should return 1 available period", func(t *testing.T) {
		cr, br := initAvailabilityServiceMocks()

		booking := domain.Booking{StayPeriod: domain.Period{
			Start: time.Date(2025, 9, 20, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
		}}

		cr.GetByNameFunc = func(_ context.Context, _ string) (domain.Cottage, error) {
			return domain.Cottage{}, nil
		}
		br.GetBookingsFunc = func(_ context.Context, _ []bson.ObjectID) ([]domain.Booking, error) {
			return []domain.Booking{booking}, nil
		}

		s := NewAvailabilityService(cr, br)
		got, _ := s.GetAvailablePeriods(context.Background(), "X", requestPeriod)

		wanted := []domain.Period{
			{
				Start: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 9, 20, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
			},
		}

		if !reflect.DeepEqual(got, wanted) {
			t.Errorf("GetAvailablePeriods() got = %v, want %v", got, wanted)
		}
	})

	t.Run("when cottage X has 1 booking matching the beginning of the period, then should return 1 available period", func(t *testing.T) {
		cr, br := initAvailabilityServiceMocks()

		booking := domain.Booking{StayPeriod: domain.Period{
			Start: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC),
		}}

		cr.GetByNameFunc = func(_ context.Context, _ string) (domain.Cottage, error) {
			return domain.Cottage{}, nil
		}
		br.GetBookingsFunc = func(_ context.Context, _ []bson.ObjectID) ([]domain.Booking, error) {
			return []domain.Booking{booking}, nil
		}

		s := NewAvailabilityService(cr, br)
		got, _ := s.GetAvailablePeriods(context.Background(), "X", requestPeriod)

		wanted := []domain.Period{
			{
				Start: time.Date(2025, 9, 11, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
			},
		}

		if !reflect.DeepEqual(got, wanted) {
			t.Errorf("GetAvailablePeriods() got = %v, want %v", got, wanted)
		}
	})

	t.Run("when cottage X has 1 booking matching the end of the period, then should return 1 available period", func(t *testing.T) {
		cr, br := initAvailabilityServiceMocks()

		booking := domain.Booking{StayPeriod: domain.Period{
			Start: time.Date(2025, 9, 20, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC),
		}}

		cr.GetByNameFunc = func(_ context.Context, _ string) (domain.Cottage, error) {
			return domain.Cottage{}, nil
		}
		br.GetBookingsFunc = func(_ context.Context, _ []bson.ObjectID) ([]domain.Booking, error) {
			return []domain.Booking{booking}, nil
		}

		s := NewAvailabilityService(cr, br)
		got, _ := s.GetAvailablePeriods(context.Background(), "X", requestPeriod)

		wanted := []domain.Period{
			{
				Start: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2025, 9, 20, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
			},
		}

		if !reflect.DeepEqual(got, wanted) {
			t.Errorf("GetAvailablePeriods() got = %v, want %v", got, wanted)
		}
	})
}

func Test_availabilityService_GetAvailablePeriodsByCottageType(t *testing.T) {
	requestPeriod := domain.Period{
		Start: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
	}

	t.Run("when Cottage A1 and A2 is available and Cottage A3 is not available, then should return availablePeriods for A1 and A2, but empty slice for A3", func(t *testing.T) {
		cr, br := initAvailabilityServiceMocks()

		bookingId := bson.NewObjectIDFromTimestamp(time.Now())
		cottageA1 := domain.Cottage{Name: "A1"}
		cottageA2 := domain.Cottage{Name: "A2"}
		cottageA3 := domain.Cottage{Name: "A3", Bookings: []bson.ObjectID{bookingId}}
		cottageType := "A1"

		booking := domain.Booking{StayPeriod: domain.Period{
			Start: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
		}}

		cr.GetByTypeFunc = func(_ context.Context, _ string) ([]domain.Cottage, error) {
			return []domain.Cottage{cottageA1, cottageA2, cottageA3}, nil
		}
		br.GetBookingsFunc = func(_ context.Context, ids []bson.ObjectID) ([]domain.Booking, error) {
			if len(ids) == 0 {
				return nil, nil
			}
			return []domain.Booking{booking}, nil
		}

		wanted := []domain.CottageAvailablePeriod{
			{
				Name: "A1",
				Periods: []domain.Period{
					{
						Start: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
					},
				},
			}, {
				Name: "A2",
				Periods: []domain.Period{
					{
						Start: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
					},
				},
			}, {
				Name:    "A3",
				Periods: []domain.Period{},
			},
		}

		s := NewAvailabilityService(cr, br)
		got, _ := s.GetAvailablePeriodsByCottageType(context.Background(), cottageType, requestPeriod)

		if !reflect.DeepEqual(got, wanted) {
			t.Errorf("GetAvailablePeriodsByCottageType() got = %v, want %v", got, wanted)
		}
	})

	t.Run("when cottage of type X is available, but requested type is Y, then should return empty slice", func(t *testing.T) {
		cr, br := initAvailabilityServiceMocks()

		cottageA1 := domain.Cottage{Name: "A1"}

		cr.GetByTypeFunc = func(_ context.Context, _ string) ([]domain.Cottage, error) {
			return []domain.Cottage{cottageA1}, nil
		}
		br.GetBookingsFunc = func(_ context.Context, _ []bson.ObjectID) ([]domain.Booking, error) {
			return nil, nil
		}

		cottageType := "typeA"

		s := NewAvailabilityService(cr, br)
		got, _ := s.GetAvailablePeriodsByCottageType(context.Background(), cottageType, requestPeriod)

		wanted := []domain.CottageAvailablePeriod{
			{
				Name: "A1",
				Periods: []domain.Period{
					{
						Start: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
					},
				},
			},
		}

		if !reflect.DeepEqual(got, wanted) {
			t.Errorf("GetAvailablePeriodsByCottageType() got = %v, want %v", got, wanted)
		}

	})
}

func Test_availabilityService_IsCottageFree(t *testing.T) {
	t.Run("when cottage has no bookings for the period, then should return true", func(t *testing.T) {
		cr, br := initAvailabilityServiceMocks()

		cottageA1 := domain.Cottage{Name: "A1"}

		cr.GetByNameFunc = func(_ context.Context, _ string) (domain.Cottage, error) {
			return cottageA1, nil
		}
		br.GetBookingsFunc = func(_ context.Context, _ []bson.ObjectID) ([]domain.Booking, error) {
			return nil, nil
		}

		cottageName := "A1"
		requestPeriod := domain.Period{
			Start: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
		}

		s := NewAvailabilityService(cr, br)
		got, _ := s.IsCottageAvailable(context.Background(), cottageName, requestPeriod)

		if got != true {
			t.Errorf("IsCottageAvailable() got = %v, want %v", got, true)
		}
	})

	t.Run("when cottage has no overlapping bookings, then should return true", func(t *testing.T) {
		cr, br := initAvailabilityServiceMocks()

		cottageA1 := domain.Cottage{Name: "A1"}
		booking := domain.Booking{StayPeriod: domain.Period{
			Start: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC),
		}}

		cr.GetByNameFunc = func(_ context.Context, _ string) (domain.Cottage, error) {
			return cottageA1, nil
		}
		br.GetBookingsFunc = func(_ context.Context, _ []bson.ObjectID) ([]domain.Booking, error) {
			return []domain.Booking{booking}, nil
		}

		cottageName := "A1"
		requestPeriod := domain.Period{
			Start: time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
		}

		s := NewAvailabilityService(cr, br)
		got, _ := s.IsCottageAvailable(context.Background(), cottageName, requestPeriod)

		if got != true {
			t.Errorf("IsCottageAvailable() got = %v, want %v", got, true)
		}
	})

	t.Run("when cottage has overlapping bookings, then should return false", func(t *testing.T) {
		cr, br := initAvailabilityServiceMocks()

		cottageA1 := domain.Cottage{Name: "A1"}
		booking := domain.Booking{StayPeriod: domain.Period{
			Start: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 10, 10, 0, 0, 0, 0, time.UTC),
		}}

		cr.GetByNameFunc = func(_ context.Context, _ string) (domain.Cottage, error) {
			return cottageA1, nil
		}
		br.GetBookingsFunc = func(_ context.Context, _ []bson.ObjectID) ([]domain.Booking, error) {
			return []domain.Booking{booking}, nil
		}

		cottageName := "A1"
		requestPeriod := domain.Period{
			Start: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond),
		}

		s := NewAvailabilityService(cr, br)
		got, _ := s.IsCottageAvailable(context.Background(), cottageName, requestPeriod)

		if got != false {
			t.Errorf("IsCottageAvailable() got = %v, want %v", got, false)
		}
	})
}
