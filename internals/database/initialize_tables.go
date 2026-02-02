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
				--create categories table
			CREATE TABLE IF NOT EXISTS categories (
    			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    			name VARCHAR(50) NOT NULL,
    			slug VARCHAR(100) NOT NULL,
    			parent_id UUID NULL,
    			level INT NOT NULL DEFAULT 1,
    			is_active BOOLEAN DEFAULT TRUE,
    			display_order INT DEFAULT 0,
    			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    			FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE CASCADE,
    			CONSTRAINT chk_max_level CHECK (level <= 5)
			);

			CREATE INDEX IF NOT EXISTS idx_categories_parent ON categories(parent_id);
			CREATE UNIQUE INDEX IF NOT EXISTS uq_root_category_slug ON categories (slug) WHERE parent_id IS NULL;

			-- Child categories (parent_id IS NOT NULL)
			CREATE UNIQUE INDEX IF NOT EXISTS uq_child_category_slug ON categories (slug, parent_id) WHERE parent_id IS NOT NULL;


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
