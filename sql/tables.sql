CREATE TABLE "ports" (
	"name" text NOT NULL,
	"version" text NOT NULL,
	"newVersion" text NOT NULL,
	"category" text NOT NULL,
	"checkedAt" timestamp,
	"updatedAt" timestamp DEFAULT CURRENT_TIMESTAMP,
	"maintainer" text NOT NULL,
	UNIQUE ("category", "name")
);

CREATE TABLE "repo" (
	"id" integer PRIMARY KEY DEFAULT 1,
	"lastCommit" text NOT NULL,
	"syncedAt" timestamp,
	CHECK ("id" = 1)
);

CREATE TABLE "hosts" (
	"hostname" text,
	"accessedAt" timestamp DEFAULT CURRENT_TIMESTAMP,
	"isDown" boolean DEFAULT FALSE NOT NULL,
	UNIQUE ("hostname")
);

INSERT INTO "repo" (
	"lastCommit",
	"syncedAt"
) VALUES (
	'',
	CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION update_updatedAt_column()
RETURNS TRIGGER AS $$
BEGIN
	-- Check if any column has changed
	IF NEW.* IS DISTINCT FROM OLD.* THEN
		NEW."updatedAt" = CURRENT_TIMESTAMP;
	ELSE
		NEW."updatedAt" = OLD."updatedAt";
	END IF;
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_ports
BEFORE UPDATE ON ports
FOR EACH ROW
EXECUTE FUNCTION update_updatedAt_column();
