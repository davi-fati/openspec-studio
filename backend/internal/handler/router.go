package handler

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/davidlima/openspec-studio/backend/docs" // registers the generated OpenAPI spec with swag
	"github.com/davidlima/openspec-studio/backend/internal/service"
)

// New builds the Gin engine with every /api route registered, plus the
// interactive Swagger UI (backed by the generated OpenAPI spec) at /swagger.
func New(
	projects *service.ProjectService,
	specs *service.SpecService,
	changes *service.ChangeService,
	projectContext *service.ContextService,
	overview *service.OverviewService,
	providers *service.ProviderService,
	specflows *service.SpecflowService,
	watcher ProjectWatcher,
	events *service.EventBroadcaster,
) *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		api.GET("/healthz", HealthzHandler)
		api.GET("/events", EventsHandler(events))
		api.GET("/overview", NewOverviewHandler(overview).Get)
		api.GET("/fs/browse", Browse(projects))

		projectsGroup := api.Group("/projects")
		{
			h := NewProjectHandler(projects, watcher)
			projectsGroup.GET("", h.List)
			projectsGroup.POST("", h.Open)
			projectsGroup.POST("/new", h.New)

			ch := NewContextHandler(projectContext)
			projectsGroup.GET("/:id/reviews", ch.ListReviews)
			projectsGroup.POST("/:id/reviews", ch.CreateReview)
			projectsGroup.DELETE("/:id/reviews/:reviewId", ch.DeleteReview)
			projectsGroup.GET("/:id/debts", ch.ListDebts)
			projectsGroup.POST("/:id/debts", ch.CreateDebt)
			projectsGroup.PATCH("/:id/debts/:debtId", ch.UpdateDebt)
			projectsGroup.DELETE("/:id/debts/:debtId", ch.DeleteDebt)
		}

		specsGroup := api.Group("/specs")
		{
			h := NewSpecHandler(specs, projects, projectContext)
			specsGroup.GET("", h.List)
			specsGroup.POST("", h.Create)
			specsGroup.POST("/generate", h.Generate)
			specsGroup.GET("/:id", h.Get)
			specsGroup.PATCH("/:id", h.Update)
			specsGroup.DELETE("/:id", h.Delete)
			specsGroup.POST("/:id/dependencies", h.AddDependency)
			specsGroup.DELETE("/:id/dependencies", h.RemoveDependency)
		}

		changesGroup := api.Group("/changes")
		{
			h := NewChangeHandler(changes, projects)
			changesGroup.GET("", h.List)
			changesGroup.GET("/:id", h.Get)
		}

		providersGroup := api.Group("/ai-providers")
		{
			h := NewProviderHandler(providers)
			providersGroup.GET("", h.List)
			providersGroup.POST("", h.Create)
			providersGroup.PATCH("/:id", h.Update)
			providersGroup.DELETE("/:id", h.Delete)
			providersGroup.POST("/:id/default", h.SetDefault)
			providersGroup.POST("/:id/health", h.HealthCheck)
		}

		specflowGroup := api.Group("/specflow/flows")
		{
			h := NewSpecflowHandler(specflows)
			specflowGroup.GET("", h.List)
			specflowGroup.POST("", h.Create)
			specflowGroup.GET("/:id", h.Get)
			specflowGroup.PATCH("/:id/reorder", h.Reorder)
			specflowGroup.DELETE("/:id", h.Delete)
			specflowGroup.POST("/:id/cancel", h.Cancel)
		}
	}

	return r
}
