CREATE TABLE "ports" (
	"name" text,
	"version" text,
	"newVersion" text,
	"category" text,
	"checkedAt" timestamp,
	"updatedAt" timestamp DEFAULT CURRENT_TIMESTAMP,
	"maintainer" text,
	UNIQUE ("category", "name")
);

CREATE TABLE "hosts" (
	"hostname" text,
	"accessedAt" timestamp DEFAULT CURRENT_TIMESTAMP,
	"isDown" boolean DEFAULT FALSE NOT NULL,
	UNIQUE ("hostname")
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
