package app

import (
	"net/http"

	"org-api/config"
	"org-api/internal/database"
	"org-api/internal/handler"
	"org-api/internal/middleware"
	"org-api/internal/repository"
	"org-api/internal/service"
	"org-api/pkg/logger"
)

type App struct {
	server *http.Server
	log    *logger.Logger
}

func New(cfg *config.Config, log *logger.Logger) (*App, error) {
	db, err := database.Connect(cfg.DSN())
	if err != nil {
		return nil, err
	}

	deptRepo := repository.NewDepartmentRepository(db)
	empRepo := repository.NewEmployeeRepository(db)

	deptSvc := service.NewDepartmentService(deptRepo, empRepo)
	empSvc := service.NewEmployeeService(deptRepo, empRepo)

	deptH := handler.NewDepartmentHandler(deptSvc, log)
	empH := handler.NewEmployeeHandler(empSvc, log)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /departments/", deptH.Create)
	mux.HandleFunc("GET /departments/{id}", deptH.Get)
	mux.HandleFunc("PATCH /departments/{id}", deptH.Update)
	mux.HandleFunc("DELETE /departments/{id}", deptH.Delete)
	mux.HandleFunc("POST /departments/{id}/employees/", empH.Create)

	handler := middleware.Recover(log)(middleware.Logging(log)(mux))

	return &App{
		server: &http.Server{
			Addr:    ":" + cfg.ServerPort,
			Handler: handler,
		},
		log: log,
	}, nil
}

func (a *App) Run() error {
	a.log.Infof("starting server on %s", a.server.Addr)
	return a.server.ListenAndServe()
}
