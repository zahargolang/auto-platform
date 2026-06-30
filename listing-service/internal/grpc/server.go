package listings_grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	core_domain "listing-service/internal/core/domain"
	core_errors "listing-service/internal/core/errors"
	"listing-service/internal/grpc/listingpb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ListingGetter interface {
	GetListingByID(ctx context.Context, id uuid.UUID) (core_domain.Listing, error)
}

type Server struct {
	listingpb.UnimplementedListingServiceServer
	svc ListingGetter
}

func NewServer(svc ListingGetter) *Server {
	return &Server{svc: svc}
}

func (s *Server) GetListing(ctx context.Context, req *listingpb.GetListingRequest) (*listingpb.GetListingResponse, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid listing id: %s", req.GetId())
	}

	listing, err := s.svc.GetListingByID(ctx, id)
	if err != nil {
		if errors.Is(err, core_errors.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "listing %s not found", id)
		}
		return nil, status.Errorf(codes.Internal, "get listing: %v", err)
	}

	return &listingpb.GetListingResponse{
		Id:     listing.ID.String(),
		UserId: listing.UserID.String(),
	}, nil
}
