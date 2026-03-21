DROP SCHEMA IF EXISTS vsds CASCADE;

CREATE SCHEMA vsds;
GRANT USAGE ON SCHEMA vsds TO edadmin, edservice, edviewer;

CREATE TABLE vsds.folders (
       id    int GENERATED ALWAYS AS IDENTITY,
       name  varchar(256) NOT NULL,
       gcpid varchar(128) NOT NULL,
       PRIMARY KEY (id),
       UNIQUE (gcpid),
       UNIQUE (name)
);
GRANT SELECT, INSERT, UPDATE, DELETE ON vsds.folders TO edservice;
GRANT SELECT ON vsds.folders TO edviewer;

-- One row per file discovered in a watched folder.
-- gcpid is the Google Drive file id; unique per folder to
-- prevent double-registration within the same scan.
CREATE TABLE vsds.spreadsheets (
       id          int          GENERATED ALWAYS AS IDENTITY,
       folderid    int          NOT NULL,
       gcpid       varchar(128) NOT NULL,
       name        varchar(256) NOT NULL,
       contenttype varchar(128) NOT NULL,
       PRIMARY KEY (id),
       UNIQUE (folderid, gcpid),
       FOREIGN KEY (folderid)
           REFERENCES vsds.folders (id) ON DELETE CASCADE
);
GRANT SELECT, INSERT, UPDATE, DELETE ON vsds.spreadsheets TO edservice;
GRANT SELECT ON vsds.spreadsheets TO edviewer;

-- One row per parseable survey unit inside a spreadsheet:
--   Google Sheets file  → one row per tab  (name = tab title)
--   CSV file            → one row          (name IS NULL)
CREATE TABLE vsds.sheets (
       id            int          GENERATED ALWAYS AS IDENTITY,
       spreadsheetid int          NOT NULL,
       name          varchar(256),
       PRIMARY KEY (id),
       UNIQUE (spreadsheetid, name),
       FOREIGN KEY (spreadsheetid)
           REFERENCES vsds.spreadsheets (id) ON DELETE CASCADE
);
-- For CSV spreadsheets there is exactly one implicit (NULL-named) sheet.
CREATE UNIQUE INDEX idx_sheets_single_null
    ON vsds.sheets (spreadsheetid)
    WHERE name IS NULL;
GRANT SELECT, INSERT, UPDATE, DELETE ON vsds.sheets TO edservice;
GRANT SELECT ON vsds.sheets TO edviewer;

CREATE TABLE vsds.projects (
       id   int         GENERATED ALWAYS AS IDENTITY,
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
       projectid int NOT NULL,
       zsample   int NOT NULL,
       PRIMARY KEY (projectid, zsample),
       FOREIGN KEY (projectid)
           REFERENCES vsds.projects (id) ON DELETE CASCADE
);
GRANT SELECT, INSERT, UPDATE, DELETE ON vsds.project_zsamples TO edservice;
GRANT SELECT ON vsds.project_zsamples TO edviewer;
INSERT INTO vsds.project_zsamples (projectid, zsample) VALUES
-- A15X
(1, -20),  (1, -70),  (1, -120), (1, -170), (1, -220), (1, -270),
(1, -320), (1, -370), (1, -420), (1, -470), (1, -520), (1, -570),
(1, -620), (1, -670), (1, -720), (1, -770), (1, -820), (1, -870),
(1, -920), (1, -970), (1, -1020),
-- DW3 Stellar Density Scans
(2, 0),   (2, 50),  (2, 100), (2, 150), (2, 200), (2, 250), (2, 300),
(2, 350), (2, 400), (2, 450), (2, 500), (2, 550), (2, 600), (2, 650),
(2, 700), (2, 750), (2, 800), (2, 850), (2, 900), (2, 950), (2, 1000),
-- DW3 Logarithmic Density Scans
(3, -250), (3, -200), (3, -150), (3, -100), (3, -80),  (3, -70),
(3, -60),  (3, -50),  (3, -40),  (3, -30),  (3, -20),  (3, -10),
(3, 0),    (3, 10),   (3, 20),   (3, 30),   (3, 40),   (3, 60),
(3, 70),   (3, 80),   (3, 100),  (3, 150),  (3, 200),  (3, 250)
;

CREATE TABLE vsds.surveys (
       id        int GENERATED ALWAYS AS IDENTITY,
       projectid int NOT NULL,
       cmdrid    int NOT NULL,
       sheetid   int NOT NULL,
       FOREIGN KEY (projectid) REFERENCES vsds.projects (id),
       FOREIGN KEY (cmdrid)    REFERENCES common.cmdrs (id),
       FOREIGN KEY (sheetid)
           REFERENCES vsds.sheets (id) ON DELETE CASCADE,
       PRIMARY KEY (id)
);
GRANT SELECT, INSERT ON vsds.surveys TO edservice;
GRANT SELECT ON vsds.surveys TO edviewer;

CREATE TABLE vsds.surveypoints (
       id          int    GENERATED ALWAYS AS IDENTITY,
       surveyid    int    NOT NULL,
       sysid       bigint NOT NULL,
       zsample     int    NOT NULL,
       syscount    int    NOT NULL,
       maxdistance real   NOT NULL,
       PRIMARY KEY (id),
       FOREIGN KEY (surveyid)
           REFERENCES vsds.surveys (id) ON DELETE CASCADE,
       FOREIGN KEY (sysid) REFERENCES common.systems (id),
       UNIQUE (surveyid, zsample),
       UNIQUE (surveyid, sysid),
       CHECK (syscount >= 0 AND syscount <= 50),
       CHECK (maxdistance > 0 AND maxdistance <= 20)
);
GRANT SELECT, INSERT ON vsds.surveypoints TO edservice;
GRANT SELECT ON vsds.surveypoints TO edviewer;

CREATE TABLE vsds.folder_processing (
       id         int       GENERATED ALWAYS AS IDENTITY,
       folderid   int       NOT NULL,
       receivedat timestamp NOT NULL DEFAULT now(),
       startedat  timestamp,
       finishedat timestamp,
       PRIMARY KEY (id),
       FOREIGN KEY (folderid)
           REFERENCES vsds.folders (id) ON DELETE CASCADE
);
GRANT SELECT, INSERT, UPDATE, DELETE ON vsds.folder_processing TO edservice;

-- Tracks per-sheet processing outcome within a folder_processing run.
-- success: NULL = in progress, true = ok, false = failed.
-- Cascades from both sides: deleting the processing run or the sheet
-- removes the tracking row.
CREATE TABLE vsds.sheet_processing (
       id      bigint GENERATED ALWAYS AS IDENTITY,
       procid  int    NOT NULL,
       sheetid int    NOT NULL,
       success boolean,
       message text,
       PRIMARY KEY (id),
       FOREIGN KEY (procid)
           REFERENCES vsds.folder_processing (id) ON DELETE CASCADE,
       FOREIGN KEY (sheetid)
           REFERENCES vsds.sheets (id) ON DELETE CASCADE
);
GRANT SELECT, INSERT, UPDATE, DELETE ON vsds.sheet_processing TO edservice;

-- Named column layouts used to recognise and parse survey sheets.
-- Each variant belongs to exactly one project.
CREATE TABLE vsds.spreadsheetvariants (
       id                int         GENERATED ALWAYS AS IDENTITY,
       projectid         int         NOT NULL,
       name              varchar(64) NOT NULL,
       headerrow         int         NOT NULL,
       sysnamecolumn     int         NOT NULL,
       zsamplecolumn     int         NOT NULL,
       systemcountcolumn int         NOT NULL,
       maxdistancecolumn int         NOT NULL,
       PRIMARY KEY (id),
       FOREIGN KEY (projectid)
           REFERENCES vsds.projects (id) ON DELETE CASCADE
);
GRANT SELECT, INSERT, UPDATE, DELETE
    ON vsds.spreadsheetvariants TO edservice;
GRANT SELECT ON vsds.spreadsheetvariants TO edviewer;

-- Individual cell-value assertions that fingerprint a variant.
-- (variantid, col, row) is unique: no two checks on the same cell.
CREATE TABLE vsds.spreadsheetvariant_checks (
       id        int         GENERATED ALWAYS AS IDENTITY,
       variantid int         NOT NULL,
       col       int         NOT NULL,
       row       int         NOT NULL,
       value     varchar(64) NOT NULL,
       PRIMARY KEY (id),
       UNIQUE (variantid, col, row),
       FOREIGN KEY (variantid)
           REFERENCES vsds.spreadsheetvariants (id) ON DELETE CASCADE
);
GRANT SELECT, INSERT, UPDATE, DELETE
    ON vsds.spreadsheetvariant_checks TO edservice;
GRANT SELECT ON vsds.spreadsheetvariant_checks TO edviewer;

-- Seed the four known variants (previously hard-coded in Go).
WITH p AS (
    SELECT id, name FROM vsds.projects
)
INSERT INTO vsds.spreadsheetvariants
    (projectid, name, headerrow,
     sysnamecolumn, zsamplecolumn,
     systemcountcolumn, maxdistancecolumn)
SELECT p.id,
       v.name, v.headerrow,
       v.sysnamecolumn, v.zsamplecolumn,
       v.systemcountcolumn, v.maxdistancecolumn
FROM (VALUES
    ('DW3 Stellar Density Scans',     'DW3 SDS',        4, 0, 1, 2, 4),
    ('DW3 Logarithmic Density Scans', 'DW3 Logarithmic', 4, 0, 1, 2, 4),
    ('A15X CW Density Scans',         'A15X A',          4, 0, 1, 2, 3),
    ('A15X CW Density Scans',         'A15X B',          5, 0, 1, 2, 3)
) AS v(projectname, name, headerrow,
       sysnamecolumn, zsamplecolumn,
       systemcountcolumn, maxdistancecolumn)
JOIN p ON p.name = v.projectname
;

-- Seed header checks for each variant.
WITH v AS (
    SELECT id, name FROM vsds.spreadsheetvariants
)
INSERT INTO vsds.spreadsheetvariant_checks (variantid, col, row, value)
SELECT v.id, c.col, c.row, c.value
FROM (VALUES
    ('DW3 SDS', 0, 4, 'System'),
    ('DW3 SDS', 2, 4, 'System Count'),
    ('DW3 SDS', 1, 5, '0'),
    ('DW3 SDS', 6, 4, 'X'),
    ('DW3 SDS', 7, 4, 'Z'),
    ('DW3 SDS', 8, 4, 'Y'),
    ('DW3 Logarithmic', 0, 4, 'System'),
    ('DW3 Logarithmic', 2, 4, 'System Count'),
    ('DW3 Logarithmic', 1, 5, '-250'),
    ('DW3 Logarithmic', 6, 4, 'X'),
    ('DW3 Logarithmic', 7, 4, 'Z'),
    ('DW3 Logarithmic', 8, 4, 'Y'),
    ('A15X A', 0, 4, 'System'),
    ('A15X A', 2, 4, 'n'),
    ('A15X A', 5, 4, 'X'),
    ('A15X A', 6, 4, 'Z'),
    ('A15X A', 7, 4, 'Y'),
    ('A15X B', 0, 5, 'System'),
    ('A15X B', 2, 5, 'n'),
    ('A15X B', 5, 5, 'X'),
    ('A15X B', 6, 5, 'Z'),
    ('A15X B', 7, 5, 'Y')
) AS c(variantname, col, row, value)
JOIN v ON v.name = c.variantname
;
