CREATE TABLE properties (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    address TEXT NOT NULL,
    city VARCHAR(50),
    type VARCHAR(20) CHECK (type IN ('Casa', 'Apartamento', 'Terreno', 'Oficina')),
    bedrooms INT DEFAULT 0,
    bathrooms INT DEFAULT 0,
    area_sqm DECIMAL(10, 2),
    lat DECIMAL(10, 8),
    lng DECIMAL(11, 8),
    main_image VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_properties_city ON properties(city);
CREATE INDEX idx_properties_price ON properties(price);
CREATE INDEX idx_properties_type ON properties(type);