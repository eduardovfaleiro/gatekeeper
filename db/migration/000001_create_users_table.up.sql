DO $$ 
BEGIN 
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'id_uuid_v7') THEN
        CREATE DOMAIN id_uuid_v7 AS uuid DEFAULT uuidv7() NOT NULL;
    END IF;
END $$;

create table "users" (
	id id_uuid_v7 primary key,
	email varchar(100) not null unique,
	password_hash text not null,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
 );