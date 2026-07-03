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
}

type Service interface {
	CreateListing(ctx context.Context, userID uuid.UUID, title, description string, price int64,
		make_, model string, year, mileage int, color string,
		bodyType core_domain.BodyType, fuelType core_domain.FuelType,
		transmission core_domain.TransmissionType, engineVolume float64,
		city, region string, photoURLs []string) (core_domain.Listing, error)

	GetListingByID(ctx context.Context, id uuid.UUID) (core_domain.Listing, error)
	GetListings(ctx context.Context, filter core_domain.ListingFilter) ([]core_domain.Listing, error)
	UpdateListing(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, update core_domain.ListingUpdate) (core_domain.Listing, error)
	DeleteListing(ctx context.Context, id uuid.UUID, requesterID uuid.UUID) error
}

func NewListingsHandler(service Service) *ListingsHandler {
	return &ListingsHandler{service: service}
}

func (h *ListingsHandler) InitRoutes(
	log *core_logger.Logger,
	allowedOrigins []string,
) *gin.Engine {
	router := gin.Default()
	// RequestID переиспользует $request_id, выставленный nginx Ingress (см.
	// helm/.../ingress.yaml), чтобы один и тот же запрос было видно под
	// одним ID и в auth-service, и здесь. Logger кладёт его в контекст,
	// чтобы достать через core_logger.FromContext в хендлерах.
	router.Use(
		core_middleware.CORS(allowedOrigins),
		core_middleware.RequestID(),
		core_middleware.Logger(log),
		core_middleware.Trace(),
		core_middleware.Panic(),
	)

	public := router.Group("/api/listings")
	{
		public.GET("/health", h.Health)
		public.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		public.GET("", h.GetListings)
		public.GET("/:id", h.GetListing)
	}

	// отдельный префикс — чтобы Ingress мог защитить эти операции через
	// auth-url по path, не пересекаясь с публичным GET /api/listings/:id
	mine := router.Group("/api/listings/mine")
	mine.Use(core_middleware.RequireUserID())
	{
		mine.POST("", h.CreateListing)
		mine.PATCH("/:id", h.UpdateListing)
		mine.DELETE("/:id", h.DeleteListing)
	}

	return router
}
