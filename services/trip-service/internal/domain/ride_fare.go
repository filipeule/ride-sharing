package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	tripTypes "ride-sharing/services/trip-service/pkg/types"
	pb "ride-sharing/shared/proto/trip"
	"time"
)

type RideFareModel struct {
	ID                primitive.ObjectID
	UserID            string
	PackageSlug       string
	TotalPriceInCents float64
	Expires           time.Time
	Route             *tripTypes.OsrmAPIResponse
}

func (r *RideFareModel) ToProto() *pb.RideFare {
	return &pb.RideFare{
		Id:                r.ID.Hex(),
		UserID:            r.UserID,
		PackageSlug:       r.PackageSlug,
		TotalPriceInCents: r.TotalPriceInCents,
	}
}

func ToRideFaresProto(fares []*RideFareModel) []*pb.RideFare {
	rideFares := make([]*pb.RideFare, 0, len(fares))

	for _, f := range fares {
		rideFares = append(rideFares, f.ToProto())
	}

	return rideFares
}
