package listings_transport_http

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	core_domain "listing-service/internal/core/domain"
	core_logger "listing-service/internal/core/logger"
	core_middleware "listing-service/internal/core/transport/http/middleware"

	_ "listing-service/docs" // swagger docs
)

type ListingsHandler struct {
	service Service
	log     *core_logger.Logger
}

type Service interface {
	CreateListing(ctx context.Context, userID uuid.UUID, title, description string, price int64,
		make_, model string, year, mileage int, color string,
		bodyType core_domain.BodyType, fuelType core_domain.FuelType,
		transmission core_domain.TransmissionType, engineVolume float64,
		city, region string) (core_domain.Listing, error)

	GetListingByID(ctx context.Context, id uuid.UUID) (core_domain.Listing, error)
	GetListings(ctx context.Context, filter core_domain.ListingFilter) ([]core_domain.Listing, error)
	UpdateListing(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, update core_domain.ListingUpdate) (core_domain.Listing, error)
	DeleteListing(ctx context.Context, id uuid.UUID, requesterID uuid.UUID) error
}

func NewListingsHandler(service Service, log *core_logger.Logger) *ListingsHandler {
	return &ListingsHandler{service: service, log: log}
}

func (h *ListingsHandler) InitRoutes() *gin.Engine {
	router := gin.Default()

	

	public := router.Group("/api/listings")
	{
		public.GET("/health", h.Health)
		public.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		public.GET("", h.GetListings)
		public.GET("/:id", h.GetListing)
	}

	authorized := router.Group("/api/listings")
	authorized.Use(core_middleware.AuthMiddleware())
	{
		authorized.POST("", h.CreateListing)
		authorized.PATCH("/:id", h.UpdateListing)
		authorized.DELETE("/:id", h.DeleteListing)
	}

	return router
}
