package schema

import (
	"database/sql"
	"fmt"
)

func InitServicesSchema(db *sql.DB) error {
	if err := createServiceTable(db); err != nil {
		return fmt.Errorf("failed to create service table: %v", err)
	}

	if err := createServicePaymentTable(db); err != nil {
		return fmt.Errorf("failed to create service payment table: %v", err)
	}

	return nil
}

func createServiceTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS service (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        description TEXT,
        provider VARCHAR(255),
        category VARCHAR(100),
        cost DECIMAL(10, 2),
        recurring_period VARCHAR(50),
        next_payment_date DATE,
        auto_renew BOOLEAN DEFAULT true,
        notification_days INTEGER DEFAULT 7,
        family_id INTEGER REFERENCES family_account(id) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE INDEX IF NOT EXISTS idx_service_family ON service(family_id);
    CREATE INDEX IF NOT EXISTS idx_service_next_payment ON service(next_payment_date);
    CREATE INDEX IF NOT EXISTS idx_service_category ON service(category);
    `
    
    _, err := db.Exec(query)
    return err
}

func createServicePaymentTable(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS service_payment (
        id SERIAL PRIMARY KEY,
        service_id INTEGER REFERENCES service(id) ON DELETE CASCADE,
        amount DECIMAL(10, 2) NOT NULL,
        payment_date DATE NOT NULL,
        status VARCHAR(50) NOT NULL,
        notes TEXT,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE INDEX IF NOT EXISTS idx_service_payment_service ON service_payment(service_id);
    CREATE INDEX IF NOT EXISTS idx_service_payment_date ON service_payment(payment_date);
    CREATE INDEX IF NOT EXISTS idx_service_payment_status ON service_payment(status);
    `
    
    _, err := db.Exec(query)
    return err
}