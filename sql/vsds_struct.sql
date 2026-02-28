DROP SCHEMA IF EXISTS vsds CASCADE;

CREATE SCHEMA vsds;
GRANT USAGE ON SCHEMA vsds TO edadmin, edservice, edviewer;

CREATE TABLE vsds.folders (
       id		 							int GENERATED ALWAYS AS IDENTITY,
			 name								varchar(256) NOT NULL,
			 gcpid						varchar(128) NOT NULL,
			 PRIMARY KEY (id),
			 UNIQUE (gcpid),
			 UNIQUE (name)
);
GRANT SELECT, INSERT, UPDATE, DELETE ON vsds.folders TO edservice;
GRANT SELECT ON vsds.folders TO edviewer;

CREATE TABLE vsds.spreadsheets (
       id		 						int GENERATED ALWAYS AS IDENTITY,
			 folderid					int NOT NULL,
			 contenttype			varchar(64) NOT NULL,
			 PRIMARY KEY (id),
			 FOREIGN KEY (folderid) REFERENCES vsds.folders (id) ON DELETE CASCADE
);
GRANT SELECT, INSERT, UPDATE, DELETE ON vsds.spreadsheets TO edservice;
GRANT SELECT ON vsds.spreadsheets TO edviewer;

CREATE TABLE vsds.projects (
       id int GENERATED ALWAYS AS IDENTITY,
       name varchar(64) NOT NULL,
       PRIMARY KEY (id)
);
GRANT SELECT, INSERT, UPDATE, DELETE ON vsds.projects TO edservice;
GRANT SELECT ON vsds.projects TO edviewer;
INSERT INTO vsds.projects (name) VALUES
('A15X CW Density Scans'),
('DW3 Stellar Density Scans'),
('DW3 Logarithmic Density Scans')
;

CREATE TABLE vsds.project_zsamples (
			 projectid				int NOT NULL,
			 zsample					int	NOT NULL,
			 PRIMARY KEY (projectid, zsample),
			 FOREIGN KEY (projectid) REFERENCES vsds.projects (id) ON DELETE CASCADE
);
GRANT SELECT, INSERT, UPDATE, DELETE ON vsds.project_zsamples TO edservice;
GRANT SELECT ON vsds.project_zsamples TO edviewer;
INSERT INTO vsds.project_zsamples (projectid, zsample) VALUES
-- A15X
(1, -20), (1, -70), (1, -120), (1, -170), (1, -220), (1, -270), (1, -320),
(1, -370), (1, -420), (1, -470), (1, -520), (1, -570), (1, -620), (1, -670),
(1, -720), (1, -770), (1, -820), (1, -870), (1, -920), (1, -970), (1, -1020),
-- DW3 SDS
(2, 0), (2, 50), (2, 100), (2, 150), (2, 200), (2, 250), (2, 300), (2, 350),
(2, 400), (2, 450), (2, 500), (2, 550), (2, 600), (2, 650), (2, 700), (2, 750),
(2, 800), (2, 850), (2, 900), (2, 950), (2, 1000),
-- DW3 Logarithmic
(3, -250), (3, -200), (3, -150), (3, -100), (3, -80), (3, -70), (3, -60),
(3, -50), (3, -40), (3, -30), (3, -20), (3, -10), (3, 0), (3, 10), (3, 20),
(3, 30), (3, 40), (3, 60), (3, 70), (3, 80), (3, 100), (3, 150), (3, 200),
(3, 250)
;

CREATE TABLE vsds.surveys (
       id 	 							int				GENERATED ALWAYS AS IDENTITY,
       projectid					int				NOT NULL,
       cmdrid							int				NOT NULL,
			 sheetid 						int				NOT NULL,
       FOREIGN KEY (projectid) REFERENCES vsds.projects (id),
       FOREIGN KEY (cmdrid) REFERENCES common.cmdrs(id),
			 FOREIGN KEY (sheetid) REFERENCES vsds.spreadsheets(id) ON DELETE CASCADE,
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
       FOREIGN KEY (surveyid) REFERENCES vsds.surveys(id) ON DELETE CASCADE,
			 FOREIGN KEY (sysid) REFERENCES common.systems(id),
       UNIQUE (surveyid, zsample),
       UNIQUE (surveyid, sysid),
       CHECK (syscount >= 0 AND syscount <= 50),
       CHECK (maxdistance > 0 AND maxdistance <= 20)
);
GRANT SELECT, INSERT ON vsds.surveypoints TO edservice;
GRANT SELECT ON vsds.surveypoints TO edviewer;

CREATE TABLE vsds.folder_processing (
       id		 						int GENERATED ALWAYS AS IDENTITY,
			 folderid					int NOT NULL,
			 receivedat				timestamp		 NOT NULL DEFAULT	now(),
			 startedat				timestamp,
			 finishedat				timestamp,
			 PRIMARY KEY (id),
			 FOREIGN KEY (folderid) REFERENCES vsds.folders (id) ON DELETE CASCADE
);
GRANT SELECT, INSERT, UPDATE, DELETE ON vsds.folder_processing TO edservice;

CREATE TABLE vsds.spreadsheet_processing (
			 id		 						bigint GENERATED ALWAYS AS IDENTITY,
			 procid						int NOT NULL,
			 sheetid					int NOT NULL,
			 success					boolean,
			 message					text,
			 PRIMARY KEY (id),
			 FOREIGN KEY (procid) REFERENCES vsds.folder_processing(id) ON DELETE CASCADE,
			 FOREIGN KEY (sheetid) REFERENCES vsds.spreadsheets (id) ON DELETE CASCADE
);
GRANT SELECT, INSERT, UPDATE, DELETE ON vsds.spreadsheet_processing TO edservice;
