# ed-survey-tools

## Getting started with CLI

To build the tool run `gmake build`.

Running the cli needs 3 things:

 1. A google serviceaccount, please see notes. Extract the service account's credentials.json, that's one of the required args
 1. An entry sheet, which is a spreadsheet, with a single sheet, where the A column has 1 entry per row. Each cell is either a link to an actual survey sheet, or just the ID of it
 1. A running postgresql database with the schema created, see example config file for connection params.

Running the cli will ingest all sheets of the referenced spreadsheets which are matching the criterias. Once cli finished, you can inspect the data in the DB. Please see the available views for example calculations, feel free to experiment.

## PostgreSQL database

Provide a functional PostgreSQL database, there are countless articles saying how to do this. Once you have this and connected to `template`, the steps are:

 1. Create 3 roles (admin, service and viewer):
    ```
	CREATE ROLE edadmin WITH LOGIN PASSWORD 'pass';
	CREATE ROLE edservice WITH LOGIN PASSWORD 'pass';
	CREATE ROLE edviewer WITH NOLOGIN;
	```
	A local cli is not using all, however the DB is prepared for the next step, where administration, the service and human querying are separated by roles. The creation scripts later are referencing all of them
 1. Create a database (`edtools` as example):
	```
	CREATE DATABASE edtools_dev WITH OWNER edadmin;
	```
 1. Import the schema:
	Start `psql` once you are in the `sql/` directory, then:
	```
	\i _all.sql
	```

# Notes

Fdev:
 - Fdev [developer zone](https://user.frontierstore.net/) to create oauth apps.
 - [Developer docs](https://user.frontierstore.net/developer/docs)
 - [Useful notes](https://github.com/Athanasius/fd-api/blob/main/docs/FrontierDevelopments-oAuth2-notes.md)

Google Cloud:
 - [Creating a service account and key](https://docs.cloud.google.com/iam/docs/keys-create-delete)
