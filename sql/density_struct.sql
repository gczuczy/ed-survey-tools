DROP SCHEMA IF EXISTS density CASCADE;

CREATE SCHEMA density;
GRANT USAGE ON SCHEMA density TO edadmin, edservice, edviewer;

CREATE TABLE density.campaigns (
       id int GENERATED ALWAYS AS IDENTITY,
       name varchar(64) NOT NULL,
       PRIMARY KEY (id)
);
GRANT SELECT, INSERT, UPDATE, DELETE ON density.campaigns TO edservice;
GRANT SELECT ON density.campaigns TO edviewer;
INSERT INTO density.campaigns (name) VALUES
('A15X CW Density Scans'),
('DW3 Stellar Density Scans'),
('DW3 Logarithmic Density Scans')
;

CREATE TABLE density.surveys (
       id    int		  GENERATED ALWAYS AS IDENTITY,
       campaignid int		  NOT NULL,
       cmdrid	 int		  NOT NULL,
       FOREIGN KEY (campaignid) REFERENCES density.campaigns (id),
       FOREIGN KEY (cmdrid) REFERENCES common.cmdrs(id),
       PRIMARY KEY (id)
);
GRANT SELECT, INSERT ON density.surveys TO edservice;
GRANT SELECT ON density.surveys TO edviewer;

CREATE TABLE density.surveypoints (
       id    int		  GENERATED ALWAYS AS IDENTITY,
       surveyid int	  NOT NULL,
			 sysid		bigint		NOT NULL,
       zsample	     int	  NOT NULL,
       syscount	     int	  NOT NULL,
       maxdistance   real	  NOT NULL,
       PRIMARY KEY (id),
       FOREIGN KEY (surveyid) REFERENCES density.surveys(id),
			 FOREIGN KEY (sysid) REFERENCES common.systems(id),
       UNIQUE (surveyid, zsample),
       UNIQUE (surveyid, sysid),
       CHECK (syscount >= 0 AND syscount <= 50),
       CHECK (maxdistance > 0 AND maxdistance <= 20)
);
GRANT SELECT, INSERT ON density.surveypoints TO edservice;
GRANT SELECT ON density.surveypoints TO edviewer;
