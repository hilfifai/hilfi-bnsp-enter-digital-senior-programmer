// internal/service/notification.service.go
package service

import (
	"api-jellyfish/internal/repository"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type INotificationService interface {
	SendLowStockAlert(ctx context.Context, productID uuid.UUID, currentStock, minStock int) error
	SendPurchaseOrderAlert(ctx context.Context, poNumber string, status string) error
	SendSalesOrderAlert(ctx context.Context, soNumber string, status string) error
	SendSystemAlert(ctx context.Context, title, message string) error
}

type NotificationService struct {
	emailService IEmailService
	smsService   ISMSService
	userRepo     repository.IUserRepository
	productRepo  repository.IProductRepository
}

func NewNotificationService(
	emailService IEmailService,
	smsService ISMSService,
	userRepo repository.IUserRepository,
	productRepo repository.IProductRepository,
) INotificationService {
	return &NotificationService{
		emailService: emailService,
		smsService:   smsService,
		userRepo:     userRepo,
		productRepo:  productRepo,
	}
}

func (s *NotificationService) SendLowStockAlert(ctx context.Context, productID uuid.UUID, currentStock, minStock int) error {
	// Get product details
	product, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return err
	}

	// Get users who should receive alerts (e.g., inventory managers)
	users, err := s.userRepo.GetUsersByRole(ctx, "INVENTORY_MANAGER")
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("Low Stock Alert: %s", product.Name)
	message := fmt.Sprintf(
		"Product %s (SKU: %s) is running low on stock.\nCurrent stock: %d\nMinimum stock: %d\nPlease reorder soon.",
		product.Name, product.SKU, currentStock, minStock,
	)

	// Send notifications to each user
	for _, user := range users {
		if user.Email != "" {
			s.emailService.SendEmail(ctx, user.Email, subject, message)
		}
		// Could also send SMS or push notifications
	}

	// Log the alert
	s.logAlert(ctx, "LOW_STOCK", productID.String(), message)

	return nil
}

func (s *NotificationService) SendPurchaseOrderAlert(ctx context.Context, poNumber string, status string) error {
	return nil
}

func (s *NotificationService) SendSalesOrderAlert(ctx context.Context, soNumber string, status string) error {
	return nil
}

func (s *NotificationService) SendSystemAlert(ctx context.Context, title, message string) error {
	return nil
}

func (s *NotificationService) logAlert(ctx context.Context, alertType, referenceID, message string) {
	// Implementation for logging alerts
}
