DROP SCHEMA IF EXISTS vsds CASCADE;

CREATE SCHEMA vsds;
GRANT USAGE ON SCHEMA vsds TO edadmin, edservice, edviewer;

CREATE TABLE vsds.campaigns (
       id int GENERATED ALWAYS AS IDENTITY,
       name varchar(64) NOT NULL,
       PRIMARY KEY (id)
);
GRANT SELECT, INSERT, UPDATE, DELETE ON vsds.campaigns TO edservice;
GRANT SELECT ON vsds.campaigns TO edviewer;
INSERT INTO vsds.campaigns (name) VALUES
('A15X CW Density Scans'),
('DW3 Stellar Density Scans'),
('DW3 Logarithmic Density Scans')
;

CREATE TABLE vsds.surveys (
       id    int		  GENERATED ALWAYS AS IDENTITY,
       campaignid int		  NOT NULL,
       cmdrid	 int		  NOT NULL,
       FOREIGN KEY (campaignid) REFERENCES vsds.campaigns (id),
       FOREIGN KEY (cmdrid) REFERENCES common.cmdrs(id),
       PRIMARY KEY (id)
);
GRANT SELECT, INSERT ON vsds.surveys TO edservice;
GRANT SELECT ON vsds.surveys TO edviewer;

CREATE TABLE vsds.surveypoints (
       id    int		  GENERATED ALWAYS AS IDENTITY,
       surveyid int	  NOT NULL,
			 sysid		bigint		NOT NULL,
       zsample	     int	  NOT NULL,
       syscount	     int	  NOT NULL,
       maxdistance   real	  NOT NULL,
       PRIMARY KEY (id),
       FOREIGN KEY (surveyid) REFERENCES vsds.surveys(id),
			 FOREIGN KEY (sysid) REFERENCES common.systems(id),
       UNIQUE (surveyid, zsample),
       UNIQUE (surveyid, sysid),
       CHECK (syscount >= 0 AND syscount <= 50),
       CHECK (maxdistance > 0 AND maxdistance <= 20)
);
GRANT SELECT, INSERT ON vsds.surveypoints TO edservice;
GRANT SELECT ON vsds.surveypoints TO edviewer;
