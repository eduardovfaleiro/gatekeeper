create table "users" (
	id uuid primary key,
	email varchar(100) not null unique,
	password_hash text not null,
    created_at TIMESTAMP WITH TIME ZONE not null
);