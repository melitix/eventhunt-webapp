-- Create the initial database schema

SET client_encoding = 'UTF8';

CREATE SCHEMA app;

CREATE EXTENSION citext SCHEMA app;
CREATE EXTENSION IF NOT EXISTS postgis SCHEMA app;

CREATE TABLE app.users (
	id				BIGSERIAL		PRIMARY KEY,
	username		varchar(15)		NOT NULL,
	password		varchar(100)	NOT NULL,
	first_name		varchar(50)		NOT NULL,
	last_name		varchar(50)		NOT NULL,
	last_active		timestamp		NOT NULL	DEFAULT CURRENT_TIMESTAMP,
	created_time	timestamp		NOT NULL	DEFAULT CURRENT_TIMESTAMP,
	updated_time	timestamp		NOT NULL	DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE app.email_addresses (
	id				BIGSERIAL	PRIMARY KEY,
	user_id			INTEGER		references app.users(id),
	the_value		app.citext	NOT NULL,
	preferred		boolean		NOT NULL,
	verified		boolean		NOT NULL	DEFAULT false,
	created_time	timestamp	NOT NULL	DEFAULT CURRENT_TIMESTAMP,
	updated_time	timestamp	NOT NULL	DEFAULT CURRENT_TIMESTAMP
);

CREATE TYPE app.token_purpose AS ENUM (
	'email-verify',
	'pw-reset',
	'write'
);

CREATE TABLE app.user_tokens (
	id				BIGSERIAL			PRIMARY KEY,
	user_id			INTEGER				references app.users(id),
	the_value		varchar(255)		NOT NULL,
	expiration		timestamp			NOT NULL,
	purpose			app.token_purpose	NOT NULL,
	created_time	timestamp			NOT NULL	DEFAULT CURRENT_TIMESTAMP,
	updated_time	timestamp			NOT NULL	DEFAULT CURRENT_TIMESTAMP
);

-- source data comes from GeoNames
-- visual: https://www.geonames.org/countries/
-- export: https://download.geonames.org/export/dump/countryInfo.txt
CREATE TABLE app.countries (
	iso_alpha2		char(2)			PRIMARY KEY,
	iso_alpha3		char(3)			NOT NULL UNIQUE,
	iso_numeric		SMALLINT		NOT NULL UNIQUE,
	name			varchar(80)		NOT NULL,
	created_time	timestamp		NOT NULL	DEFAULT CURRENT_TIMESTAMP,
	updated_time	timestamp		NOT NULL	DEFAULT CURRENT_TIMESTAMP
);

-- source data comes from GeoNames
-- export: https://download.geonames.org/export/dump/admin1CodesASCII.txt
CREATE TABLE app.spr (
	iso_alpha2		char(2)			NOT NULL references app.countries(iso_alpha2),
	admin1			varchar(20)		NOT NULL,
	name			varchar(80)		NOT NULL,
	created_time	timestamp		NOT NULL	DEFAULT CURRENT_TIMESTAMP,
	updated_time	timestamp		NOT NULL	DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT spr_pk PRIMARY KEY (iso_alpha2, admin1)
);

-- source data comes from GeoNames
-- export: https://download.geonames.org/export/dump/timeZones.txt
CREATE TABLE app.timezones (
	name			varchar(40)		PRIMARY KEY,
	iso_alpha2		char(2)			references app.countries(iso_alpha2),
	gmt_offset		NUMERIC(3,1)	NOT NULL,
	dst_offset		NUMERIC(3,1)	NOT NULL,
	raw_offset		NUMERIC(3,1)	NOT NULL,
	created_time	timestamp		NOT NULL	DEFAULT CURRENT_TIMESTAMP,
	updated_time	timestamp		NOT NULL	DEFAULT CURRENT_TIMESTAMP
);

-- source data comes from GeoNames
-- export: for small # of cities: https://download.geonames.org/export/dump/cities15000.zip
-- export: for medium # of cities: https://download.geonames.org/export/dump/cities5000.zip
-- export: for large # of cities: https://download.geonames.org/export/dump/cities1000.zip
CREATE TABLE app.cities (
	id				INTEGER						PRIMARY KEY,
	iso_alpha2		char(2)						references app.countries(iso_alpha2),
	admin1			varchar(20)					NOT NULL,
	name			varchar(200)				NOT NULL,
	location		geography(POINT,4326)	NOT NULL,
	timezone		varchar(40)					references app.timezones(name),
	created_time	timestamp					NOT NULL	DEFAULT CURRENT_TIMESTAMP,
	updated_time	timestamp					NOT NULL	DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT cities_fk FOREIGN KEY (iso_alpha2, admin1) references app.spr(iso_alpha2, admin1)
);

CREATE TABLE app.groups (
	id				BIGSERIAL		PRIMARY KEY,
	user_id			INTEGER			NOT NULL references app.users(id),
	name			varchar(80)		NOT NULL,
	summary			varchar(255)	NOT NULL,
	description		varchar(7000)	NOT NULL,
	slug			varchar(50)		NOT NULL,
	web_url			varchar(1000),
	city_id			INTEGER			NOT NULL references app.cities(id),
	is_private		boolean			NOT NULL,
	created_time	timestamp		NOT NULL	DEFAULT CURRENT_TIMESTAMP,
	updated_time	timestamp		NOT NULL	DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE app.venues (
	id				BIGSERIAL			PRIMARY KEY,
	name			varchar(80)			NOT NULL,
	address			varchar(255)		NOT NULL,
	city_id			INTEGER				NOT NULL references app.cities(id),
	web_url			varchar(1000)		NOT NULL,
	capacity		INTEGER				NOT NULL,
	created_time	timestamp			NOT NULL	DEFAULT CURRENT_TIMESTAMP,
	updated_time	timestamp			NOT NULL	DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE app.events (
	id				BIGSERIAL		PRIMARY KEY,
	group_id		INTEGER			references app.groups(id),
	name			varchar(80)		NOT NULL,
	start_time		timestamp		NOT NULL,
	end_time		timestamp		NOT NULL,
	summary			varchar(255)	NOT NULL DEFAULT '',
	description		varchar(7000)	NOT NULL DEFAULT '',
	web_url			varchar(1000)	NOT NULL DEFAULT '',
	announce_url	varchar(1000)	NOT NULL DEFAULT '',
	attendee_limit	INTEGER			NOT NULL DEFAULT 0,
	venue_id		INTEGER			references app.venues(id) DEFAULT null,
	created_time	timestamp		NOT NULL	DEFAULT CURRENT_TIMESTAMP,
	updated_time	timestamp		NOT NULL	DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE app.user_settings (
	id				BIGSERIAL			PRIMARY KEY,
	user_id			INTEGER				references app.users(id),
	created_time	timestamp			NOT NULL	DEFAULT CURRENT_TIMESTAMP,
	updated_time	timestamp			NOT NULL	DEFAULT CURRENT_TIMESTAMP
);

CREATE TYPE rsvp_status AS ENUM ('yes', 'in-person', 'online', 'maybe', 'no');
CREATE TYPE rsvp_role AS ENUM ('attendee', 'host', 'crew');

CREATE TABLE app.rsvps (
	id				BIGSERIAL		NOT NULL,
	event_id		BIGINT			references app.events(id),
	user_id			BIGINT			references app.users(id),
	intent			rsvp_status		NOT NULL,
	actual			rsvp_status,
	role			rsvp_role		NOT NULL,
	reminded_time	timestamp,
	created_time	timestamp		NOT NULL	DEFAULT CURRENT_TIMESTAMP,
	updated_time	timestamp		NOT NULL	DEFAULT CURRENT_TIMESTAMP,

	CONSTRAINT rsvps_pk PRIMARY KEY (event_id, user_id)
);

---- create above / drop below ----

DROP SCHEMA IF EXISTS app CASCADE;
