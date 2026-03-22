package main

import (
	"log"

	"github.com/parassoni1899/appointment-booking/internal/config"
	"github.com/parassoni1899/appointment-booking/internal/handler"
	"github.com/parassoni1899/appointment-booking/internal/repository"
	"github.com/parassoni1899/appointment-booking/internal/server"
	"github.com/parassoni1899/appointment-booking/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Load configuration
	cfg := config.LoadConfig()

	// 2. Setup database connection
	db, err := gorm.Open(postgres.Open(cfg.DBUrl), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Initialize repository, service, handler
	repo := repository.NewRepository(db)
	
	// AutoMigrate schemas and seed dummy users/coaches
	if err := repo.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database schemas: %v", err)
	}

	svc := service.NewService(repo)
	h := handler.NewHandler(svc)

	// 4. Setup router
	r := server.NewRouter(h)

	// 5. Start server
	log.Printf("Starting server on port %s...", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
