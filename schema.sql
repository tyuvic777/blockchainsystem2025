-- PostgreSQL schema for MediNet with sharding and role-specific access

-- Enable extension for sharding
CREATE EXTENSION IF NOT EXISTS citus;

-- Create base tables (sharded by user_id % 10 for 10 shards)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(120) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'doctor', 'patient')),
    name VARCHAR(120),
    blockchain_address VARCHAR(255),
    shard_id INT GENERATED ALWAYS AS (MOD(id, 10)) STORED
) PARTITION BY HASH (shard_id);

-- Create shards for users (10 shards)
DO $$
BEGIN
    FOR i IN 0..9 LOOP
        EXECUTE format('CREATE TABLE users_%s PARTITION OF users FOR VALUES WITH (MODULUS 10, REMAINDER %s)', i, i);
    END LOOP;
END;
$$;

CREATE TABLE patients (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    name VARCHAR(120) NOT NULL,
    age INT,
    gender VARCHAR(20),
    blood_type VARCHAR(10),
    medical_condition VARCHAR(255),
    date_of_admission TIMESTAMP,
    doctor VARCHAR(120),
    discharge_date TIMESTAMP,
    medication VARCHAR(255),
    test_results TEXT,
    wearable_data TEXT,
    care_plan TEXT,
    credentials TEXT,
    priority VARCHAR(20) CHECK (priority IN ('Low', 'Medium', 'High')),
    blockchain_hash VARCHAR(255),
    ipfs_hash VARCHAR(255),
    shard_id INT GENERATED ALWAYS AS (MOD(user_id, 10)) STORED
) PARTITION BY HASH (shard_id);

-- Create shards for patients (10 shards)
DO $$
BEGIN
    FOR i IN 0..9 LOOP
        EXECUTE format('CREATE TABLE patients_%s PARTITION OF patients FOR VALUES WITH (MODULUS 10, REMAINDER %s)', i, i);
    END LOOP;
END;
$$;

CREATE TABLE appointments (
    id SERIAL PRIMARY KEY,
    patient_id INT NOT NULL REFERENCES patients(id),
    patient_name VARCHAR(120) NOT NULL,
    doctor_name VARCHAR(120) NOT NULL,
    date TIMESTAMP NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('Pending', 'Approved', 'Rejected')),
    comment TEXT,
    blockchain_hash VARCHAR(255),
    shard_id INT GENERATED ALWAYS AS (MOD(patient_id, 10)) STORED
) PARTITION BY HASH (shard_id);

-- Create shards for appointments (10 shards)
DO $$
BEGIN
    FOR i IN 0..9 LOOP
        EXECUTE format('CREATE TABLE appointments_%s PARTITION OF appointments FOR VALUES WITH (MODULUS 10, REMAINDER %s)', i, i);
    END LOOP;
END;
$$;

-- Create indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_patients_user_id ON patients(user_id);
CREATE INDEX idx_appointments_patient_id ON appointments(patient_id);
CREATE INDEX idx_appointments_date ON appointments(date);

-- Grant permissions based on role (admin, doctor, patient)
GRANT SELECT, INSERT, UPDATE, DELETE ON users, patients, appointments TO admin_role;
GRANT SELECT, INSERT, UPDATE ON patients, appointments TO doctor_role;
GRANT SELECT ON patients, appointments TO patient_role;

-- Create roles (simplified for demonstration)
CREATE ROLE admin_role;
CREATE ROLE doctor_role;
CREATE ROLE patient_role;

-- Assign users to roles based on role column
DO $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN SELECT id, role FROM users LOOP
        IF r.role = 'admin' THEN
            EXECUTE format('GRANT admin_role TO user_%s', r.id);
        ELSIF r.role = 'doctor' THEN
            EXECUTE format('GRANT doctor_role TO user_%s', r.id);
        ELSE
            EXECUTE format('GRANT patient_role TO user_%s', r.id);
        END IF;
    END LOOP;
END;
$$;