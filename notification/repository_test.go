package notification_test

import (
	"api/database"
	"api/notification"
	"context"
	"testing"
	"time"
)

func TestNotificationRepository(t *testing.T) {
	conn, err := database.GetConnection(database.SQLITE, "../test.db")
	if err != nil {
		t.Fatalf("could not open database: %s", err)
	}

	tx, err := conn.DB.Begin()
	if err != nil {
		t.Fatalf("could not start transaction: %s", err)
	}

	defer tx.Rollback()

	if _, err := tx.Exec(`
        INSERT INTO companies (id, name, email, password) VALUES
        (1, "Foo", "bar", "baz"), (2, "Bar", "baz", "foo")
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if _, err := tx.Exec(`
        INSERT INTO notifications (id, company_id, message, created_at) VALUES
        (1, 1, "hi there", "2024-01-08 21:20:50"),
        (2, 1, "hi there", "2024-01-08 21:20:51"),
        (3, 1, "hi there", "2024-01-08 21:20:52"),
        (4, 1, "hi there", "2024-01-08 21:20:53"),
        (5, 1, "hi there", "2024-01-08 21:20:54"),
        (6, 1, "hi there", "2024-01-08 21:20:55"),
        (7, 1, "hi there", "2024-01-08 21:20:56"),
        (8, 1, "hi there", "2024-01-08 21:20:57"),
        (9, 1, "hi there", "2024-01-08 21:20:58"),
        (10, 1, "hi there", "2024-01-08 21:20:59"),
        (11, 1, "hi there", "2024-01-08 21:21:50"),
        (12, 1, "hi there", "2024-01-08 21:22:50"),
        (13, 1, "hi there", "2024-01-08 21:23:50"),
        (14, 1, "hi there", "2024-01-08 21:24:50"),
        (15, 1, "hi there", "2024-01-08 21:25:50"),
        (16, 1, "hi there", "2024-01-08 21:26:50"),
        (17, 1, "hi there", "2024-01-08 21:27:50"),
        (18, 1, "hi there", "2024-01-08 21:28:50"),
        (19, 1, "hi there", "2024-01-08 21:29:50"),
        (20, 1, "hi there", "2024-01-08 21:30:50"),
        (21, 2, "hi there", "2024-01-08 21:31:50"),
        (22, 2, "hi there", "2024-01-08 21:32:50"),
        (23, 2, "hi there", "2024-01-08 21:33:50"),
        (24, 2, "hi there", "2024-01-08 21:34:50"),
        (25, NULL, "broadcast", "2024-01-08 21:35:50")
    `); err != nil {
		t.Fatalf("could not seed database: %s", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("could not commit transaction: %s", err)
	}

	t.Cleanup(func() {
		if _, err := conn.DB.Exec(`DELETE FROM notifications`); err != nil {
			t.Errorf("could not clean up database: %s", err)
		}
		if _, err := conn.DB.Exec(`DELETE FROM companies`); err != nil {
			t.Errorf("could not clean up database: %s", err)
		}
	})

	repository := notification.NewRepository(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	t.Run("GetNotifications", func(t *testing.T) {
		notifications, err := repository.GetNotifications(ctx, 1)
		if err != nil {
			t.Fatalf("could not get notifications: %s", err)
		}

		if len(notifications) != 20 {
			t.Errorf("expected %d notifications, got %d", 3, len(notifications))
		}

		if notifications[0].Id != 25 {
			t.Errorf("expected id %d, got %d", 25, notifications[0].Id)
		}

		notifications, err = repository.GetNotifications(ctx, 2)
		if err != nil {
			t.Fatalf("could not get notifications: %s", err)
		}

		if len(notifications) != 5 {
			t.Errorf("expected %d notifications, got %d", 5, len(notifications))
		}
	})

	t.Run("SaveNotification", func(t *testing.T) {
		companyId := int64(1)

		notification, err := repository.SaveNotification(ctx, &notification.Notification{
			CompanyId: &companyId,
			Message:   "hello there",
		})

		if err != nil {
			t.Fatalf("could not save notification: %s", err)
		}

		if notification.Id == 0 {
			t.Error("should set notification id")
		}
	})
}
