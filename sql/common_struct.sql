DROP SCHEMA IF EXISTS common CASCADE;

CREATE SCHEMA common;
GRANT USAGE ON SCHEMA common TO edadmin, edservice, edviewer;

CREATE TABLE common.cmdrs (
       id    int		  GENERATED ALWAYS AS IDENTITY,
       name  varchar(64)	  NOT NULL UNIQUE,
			 customerid bigint,
			 isowner		boolean	NOT NULL DEFAULT false,
			 isadmin		boolean	NOT NULL DEFAULT false,
       PRIMARY KEY (id),
			 UNIQUE (customerid)
);
GRANT SELECT, INSERT, UPDATE ON common.cmdrs TO edservice;
GRANT SELECT ON common.cmdrs TO edviewer;

CREATE TABLE common.systems (
			 id		 				bigint GENERATED ALWAYS AS IDENTITY,
			 edsmid				bigint NOT NULL UNIQUE,
			 name					varchar(64)	UNIQUE,
			 x						real				NOT NULL,
			 y						real				NOT NULL,
			 z						real				NOT NULL,
			 PRIMARY KEY (id)
);
GRANT SELECT, INSERT, UPDATE ON common.systems TO edservice;
GRANT SELECT ON common.systems TO edviewer;
