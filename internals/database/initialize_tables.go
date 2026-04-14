package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var postgressSchemaOrderedTables = []struct {
	Name   string
	Schema string
}{
	{
		Name: "users",
		Schema: `
			-- Enable pgcrypto extension for UUID generation
			CREATE EXTENSION IF NOT EXISTS "pgcrypto";

			-- Create enums if they don't exist
			DO $$ 
			BEGIN 
				IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
					CREATE TYPE user_role AS ENUM ('cashier', 'admin', 'waiter', 'chef', 'delivery_staff', 'manager','customer');
				END IF;

				IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'gender_type') THEN
					CREATE TYPE gender_type AS ENUM ('male', 'female', 'other');
				END IF;

			END $$;

			--create users table

			CREATE TABLE IF NOT EXISTS users (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				email VARCHAR(100) UNIQUE NOT NULL,
				gender gender_type NOT NULL,
				image TEXT,
				is_active BOOLEAN NOT NULL DEFAULT TRUE,
				last_password_reset_at BIGINT NOT NULL,
				role user_role NOT NULL DEFAULT 'customer',
				name VARCHAR(100) NOT NULL,
				phone VARCHAR(20) UNIQUE NOT NULL,
				password TEXT NOT NULL,
				salary NUMERIC(10,2) DEFAULT 0,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			);

			--Indexes for performance optimization
			CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
			CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
			CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
			CREATE INDEX IF NOT EXISTS idx_users_fulltext
			ON users
			USING GIN (
			    to_tsvector('english', coalesce(name,'') || ' ' || coalesce(email,'') || ' ' || coalesce(phone,''))
			);
		`,
	},
	{
		Name: "raw_materials",
		Schema: `
			CREATE EXTENSION IF NOT EXISTS "pgcrypto";

			--create raw material user
			CREATE TABLE IF NOT EXISTS raw_material (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				name VARCHAR(100)  NOT NULL,
				price NUMERIC(10,2) DEFAULT 0,
				quantity NUMERIC(10,2) DEFAULT 0,
				unit VARCHAR(20) DEFAULT 'kg',
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			);

			--Indexes for performance 
			CREATE INDEX IF NOT EXISTS idx_raw_material_created_at  ON raw_material(created_at);
			CREATE INDEX IF NOT EXISTS idx_raw_material_name ON raw_material(name);

		
		`,
	},
	{
		Name: "categories",
		Schema: `
			CREATE TABLE IF NOT EXISTS categories (
    			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    			name VARCHAR(50) NOT NULL,
    			slug VARCHAR(100) NOT NULL UNIQUE,
    			is_active BOOLEAN DEFAULT TRUE,
    			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
			
			);

		`,
	},

	{
		Name: "menu_items",
		Schema: `
			CREATE TABLE IF NOT EXISTS  menu_items (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				name VARCHAR(100) NOT NULL,
				description TEXT,
				price DECIMAL(10,2) NOT NULL,
				category_id UUID NOT NULL,
				is_available BOOLEAN DEFAULT TRUE,
				image_url VARCHAR(500),
				display_order INT DEFAULT 0,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
			);

			CREATE  INDEX IF NOT EXISTS idx_menu_items_category ON menu_items(category_id);

		`,
	},
	{
		Name: "attendance",
		Schema: `
		-- Create ENUM type first
		DO $$ BEGIN
			CREATE TYPE attendance_status AS ENUM (
				'present', 
				'absent', 
				'late', 
				'half_day', 
				'leave'
			);
		EXCEPTION
			WHEN duplicate_object THEN NULL;
		END $$;

		-- Then create the table
		CREATE TABLE IF NOT EXISTS attendance (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			employee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			work_date DATE NOT NULL,
			check_in_time TIMESTAMPTZ,
			check_out_time TIMESTAMPTZ,
			need_review BOOLEAN DEFAULT FALSE,
			status attendance_status NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT uq_attendance_employee_date UNIQUE (employee_id, work_date)
		);

		-- Composite index (better than separate ones)
		CREATE INDEX IF NOT EXISTS idx_attendance_employee_date 
		ON attendance(employee_id, work_date);

		CREATE INDEX IF NOT EXISTS idx_attendance_work_date ON attendance(work_date);
	`,
	},
	{
		Name: "attendance_leave",
		Schema: `
    DO $$ BEGIN
        CREATE TYPE leave_status AS ENUM ('pending', 'approved', 'rejected');
    EXCEPTION
        WHEN duplicate_object THEN NULL;
    END $$;

    CREATE TABLE IF NOT EXISTS attendance_leave (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        employee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        checked_by UUID REFERENCES users(id) ON DELETE SET NULL,
        start_date TIMESTAMPTZ NOT NULL,
        end_date TIMESTAMPTZ NOT NULL,
        message TEXT NOT NULL,
        supervisor_message TEXT,
        status leave_status NOT NULL DEFAULT 'pending',
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        
        CONSTRAINT chk_leave_dates CHECK (end_date >= start_date),
        -- Prevent multiple leaves starting on the same day
        CONSTRAINT uq_employee_start_date UNIQUE (employee_id, start_date)
    );
    `,
	},
	{
		Name: "table_status",
		Schema: `
    DO $$ BEGIN
        CREATE TYPE table_state AS ENUM ('occupied', 'empty', 'booked');
    EXCEPTION
        WHEN duplicate_object THEN NULL;
    END $$;

    CREATE TABLE IF NOT EXISTS table_status (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        table_number INT NOT NULL,
        status table_state NOT NULL DEFAULT 'empty',
        capacity INT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        CONSTRAINT uq_table_number UNIQUE (table_number)
    );

    -- Index for faster lookup by table_number
    CREATE INDEX IF NOT EXISTS idx_table_number ON table_status(table_number);
  `,
	},
	{
		Name: "table_session",
		Schema: `

    -- Table to track active sessions per table
    CREATE TABLE IF NOT EXISTS table_session (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      table_number INT NOT NULL,
      open_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      close_time TIMESTAMPTZ,
      status table_state NOT NULL DEFAULT 'empty',
	  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

    -- Indexes for performance
    CREATE UNIQUE INDEX IF NOT EXISTS unique_active_table
ON table_session(table_number)
WHERE close_time IS NULL;
   
  `,
	},
	{
		Name: "table_validation",
		Schema: `
			CREATE TABLE IF NOT EXISTS table_validation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_number INT NOT NULL,
    phone_number TEXT NOT NULL,
    waiter_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 
    -- Composite unique constraint ensures uniqueness and automatically creates an index
    CONSTRAINT uq_table_phone UNIQUE (table_number, phone_number)
	);



		`,
	},

	{
		Name: "orders",
		Schema: `

    -- Create enum type for order status
    DO $$
    BEGIN
      IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'order_status_enum') THEN
        CREATE TYPE order_status_enum AS ENUM (
          'not-approved',
		  'approved',
          'progress',
          'completed',
          'cancelled'
        );
      END IF;
    END$$;

    -- Orders placed in a table session
    CREATE TABLE IF NOT EXISTS orders (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      table_session_id UUID REFERENCES table_session(id) ,
      customer_name VARCHAR(100),
      customer_phone VARCHAR(20),
	  waiter_id UUID REFERENCES users(id),
	   note TEXT,
      status order_status_enum DEFAULT 'not-approved',
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

    -- Indexes for performance
    CREATE INDEX IF NOT EXISTS idx_orders_table_session_id
      ON orders(table_session_id);


  `,
	},
	{
		Name: "order_items",
		Schema: `

    -- Items within an order
    CREATE TABLE IF NOT EXISTS order_items (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
      menu_item_id UUID NOT NULL,
	  status order_status_enum DEFAULT 'not-approved',
      quantity NUMERIC(10,2) DEFAULT 1,
      price NUMERIC(10,2) NOT NULL,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

    -- Indexes for performance
    CREATE INDEX IF NOT EXISTS idx_order_items_order_id
      ON order_items(order_id);
    CREATE INDEX IF NOT EXISTS idx_order_items_menu_item_id
      ON order_items(menu_item_id);
  `,
	},
	{
		Name: "payments",
		Schema: `
	-- ENUM for payment method
	DO $$ BEGIN
		CREATE TYPE payment_method_enum AS ENUM ('cash', 'online');
	EXCEPTION
		WHEN duplicate_object THEN NULL;
	END $$;

	-- ENUM for online gateways
	DO $$ BEGIN
		CREATE TYPE online_gateway_enum AS ENUM ('esewa', 'khalti', 'fonepay', 'banking', 'other');
	EXCEPTION
		WHEN duplicate_object THEN NULL;
	END $$;

	CREATE TABLE IF NOT EXISTS payments (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

		order_id UUID NOT NULL REFERENCES orders(id) ON DELETE SET NULL,

		payment_method payment_method_enum NOT NULL,
		online_gateway online_gateway_enum,

		paid_amount NUMERIC(10,2) NOT NULL CHECK (paid_amount >= 0),
		discount NUMERIC(10,2) DEFAULT 0 CHECK (discount >= 0),

		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

		-- ✅ Business rule enforcement
		CONSTRAINT chk_online_gateway_required
		CHECK (
			(payment_method = 'online' AND online_gateway IS NOT NULL)
			OR
			(payment_method = 'cash' AND online_gateway IS NULL)
		)
	);

	-- Indexes for performance
	CREATE INDEX IF NOT EXISTS idx_payments_order_id 
	ON payments(order_id);

	CREATE INDEX IF NOT EXISTS idx_payments_created_at 
	ON payments(created_at);

	CREATE INDEX IF NOT EXISTS idx_payments_method 
	ON payments(payment_method);
	`,
	}, {
		Name: "user_tokens",
		Schema: `
	CREATE TABLE IF NOT EXISTS user_tokens (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

		phone_number VARCHAR(20) NOT NULL UNIQUE,

		total_tokens NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (total_tokens >= 0),

		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	-- Index for fast lookup
	CREATE INDEX IF NOT EXISTS idx_user_tokens_phone 
	ON user_tokens(phone_number);

	CREATE INDEX IF NOT EXISTS idx_user_tokens_created_at 
	ON user_tokens(created_at);
	`,
	},
	{
		Name: "token_transactions",
		Schema: `
			CREATE TABLE IF NOT EXISTS token_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    phone_number VARCHAR(20) NOT NULL,

    amount NUMERIC(10,2) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('EARN', 'SPEND', 'STREAK')),
    source VARCHAR(50), -- 'ORDER', 'STREAK'

    reference_id UUID, -- order_id (optional)

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_token_transactions_phone 
ON token_transactions(phone_number);

CREATE INDEX IF NOT EXISTS idx_token_transactions_created_at 
ON token_transactions(created_at);
		`,
	},
	{
		Name: "CustomerStreak",
		Schema: `
			CREATE TABLE IF NOT EXISTS customer_streaks (
    phone_number VARCHAR(20) PRIMARY KEY,

    current_streak INT NOT NULL DEFAULT 0,
    last_visit DATE,
    monthly_days INT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_customer_streaks_last_visit 
ON customer_streaks(last_visit);
		`,
	},
	{
		Name: "restaurant_information",
		Schema: `

    -- Restaurant Basic Information
    CREATE TABLE IF NOT EXISTS restaurant_information (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		singleton_key BOOLEAN NOT NULL DEFAULT TRUE UNIQUE,
      name TEXT NOT NULL,
      slogan TEXT,
      logo_url TEXT,

      phone TEXT,
      email TEXT,

      address TEXT,
      country TEXT,
      state TEXT,
      city TEXT,

      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

    -- Indexes
   CREATE INDEX IF NOT EXISTS idx_restaurant_information_name
ON restaurant_information(name);

  `,
	},
	{
		Name: "password_reset_requests",
		Schema: `

		CREATE TABLE IF NOT EXISTS password_reset_requests (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

			email VARCHAR(100) NOT NULL,

			session_token TEXT NOT NULL UNIQUE,

			pin_code VARCHAR(6) NOT NULL,

			is_used BOOLEAN NOT NULL DEFAULT FALSE,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE UNIQUE INDEX IF NOT EXISTS uq_pwd_reset_active_email
		ON password_reset_requests(email)
		WHERE is_used = FALSE;

		-- Composite index for lookup (IMPORTANT for your flow)
		CREATE INDEX IF NOT EXISTS idx_pwd_reset_email_session
		ON password_reset_requests(email, session_token);

		-- Optional: faster cleanup / scanning old requests
		CREATE INDEX IF NOT EXISTS idx_pwd_reset_created_at
		ON password_reset_requests(created_at);

		
	`,
	},
}

func CreatePostgresTables(ctx context.Context, postgresPool *pgxpool.Pool) error {
	for _, t := range postgressSchemaOrderedTables {
		if err := ValidateTableName(t.Name); err != nil {
			return fmt.Errorf("invalid table name: %s", t.Name)
		}

		if _, err := postgresPool.Exec(ctx, t.Schema); err != nil {
			return fmt.Errorf("failed to create table %s: %w", t.Name, err)
		}
	}
	if err := seedRestaurantInformation(ctx, postgresPool); err != nil {
		return fmt.Errorf("failed to seed restaurant information: %w", err)
	}
	return nil
}

func ValidateTableName(tableName string) error {
	if tableName == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	if len(tableName) > 63 {
		return fmt.Errorf("table name too long (max 63 characters)")
	}

	if !pgIdentifierRegex.MatchString(tableName) {
		return fmt.Errorf("invalid table name: must start with letter or underscore, contain only letters, numbers, and underscores")
	}

	return nil
}

func seedRestaurantInformation(ctx context.Context, db *pgxpool.Pool) error {

	query := `
		INSERT INTO restaurant_information (
			name,
			slogan,
			logo_url,
			phone,
			email,
			address,
			country,
			state,
			city
		)
		SELECT
			'My Restaurant',
			'Best Food in Town',
			NULL,
			NULL,
			NULL,
			'Kathmandu',
			'Nepal',
			'Bagmati',
			'Kathmandu'
		WHERE NOT EXISTS (
			SELECT 1 FROM restaurant_information
		);
	`

	_, err := db.Exec(ctx, query)
	return err
}
