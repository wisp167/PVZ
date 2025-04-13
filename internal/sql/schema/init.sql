-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table for authentication
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('employee', 'moderator')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- PVZ (Pickup Points) table
CREATE TABLE pvz (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    registration_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    city VARCHAR(50) NOT NULL CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Receptions table
CREATE TABLE receptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    pvz_id UUID NOT NULL REFERENCES pvz(id),
    status VARCHAR(20) NOT NULL CHECK (status IN ('in_progress', 'close')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Products table with LIFO optimization
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    type VARCHAR(20) NOT NULL CHECK (type IN ('электроника', 'одежда', 'обувь')),
    reception_id UUID NOT NULL REFERENCES receptions(id),
    sequence BIGSERIAL NOT NULL,  -- For optimized LIFO operations
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance optimization
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_pvz_city ON pvz(city);
CREATE INDEX idx_receptions_pvz_id ON receptions(pvz_id);
CREATE INDEX idx_receptions_status ON receptions(status);
CREATE INDEX idx_receptions_date_time ON receptions(date_time);
CREATE INDEX idx_products_reception_id ON products(reception_id);
CREATE INDEX idx_products_date_time ON products(date_time);
CREATE INDEX idx_products_reception_sequence ON products(reception_id, sequence DESC);

-- Composite indexes for common query patterns
CREATE INDEX idx_receptions_pvz_status ON receptions(pvz_id, status);
