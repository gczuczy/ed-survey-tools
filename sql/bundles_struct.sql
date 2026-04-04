DROP SCHEMA IF EXISTS bundles CASCADE;

CREATE SCHEMA bundles;
GRANT USAGE ON SCHEMA bundles
    TO edadmin, edservice, edviewer;

CREATE TABLE bundles.bundles (
    id              int          GENERATED ALWAYS AS IDENTITY,
    measurementtype varchar(32)  NOT NULL,
    name            varchar(128) NOT NULL,
    filename        varchar(256) NOT NULL,
    generatedat     timestamp,
    autoregen       bool         NOT NULL DEFAULT false,
    status          varchar(16)  NOT NULL DEFAULT 'pending',
    errormessage    text,
    PRIMARY KEY (id),
    UNIQUE (filename),
    CHECK (status IN (
        'pending', 'queued', 'generating', 'ready', 'error'
    ))
);
GRANT SELECT, INSERT, UPDATE, DELETE
    ON bundles.bundles TO edservice;
GRANT SELECT ON bundles.bundles TO edviewer;

CREATE TABLE bundles.vsds_bundles (
    bundleid    int         NOT NULL,
    subtype     varchar(32) NOT NULL,
    allprojects bool        NOT NULL DEFAULT false,
    PRIMARY KEY (bundleid),
    FOREIGN KEY (bundleid)
        REFERENCES bundles.bundles (id) ON DELETE CASCADE,
    CHECK (subtype IN ('surveypoints', 'surveys'))
);
GRANT SELECT, INSERT, UPDATE, DELETE
    ON bundles.vsds_bundles TO edservice;
GRANT SELECT ON bundles.vsds_bundles TO edviewer;

CREATE TABLE bundles.vsds_bundle_projects (
    bundleid  int NOT NULL,
    projectid int NOT NULL,
    PRIMARY KEY (bundleid, projectid),
    FOREIGN KEY (bundleid)
        REFERENCES bundles.vsds_bundles (bundleid)
        ON DELETE CASCADE,
    FOREIGN KEY (projectid)
        REFERENCES vsds.projects (id) ON DELETE CASCADE
);
GRANT SELECT, INSERT, UPDATE, DELETE
    ON bundles.vsds_bundle_projects TO edservice;
GRANT SELECT ON bundles.vsds_bundle_projects TO edviewer;
