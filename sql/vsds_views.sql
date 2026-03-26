CREATE OR REPLACE VIEW vsds.v_surveypoints AS
WITH adjusted AS (
SELECT sp.id, sp.surveyid, s.name AS sysname,
       sp.zsample, s.x, s.y, s.z,
       sp.syscount + 1 AS corrected_n,
       greatest(least(sp.maxdistance, 20), 1) AS maxdistance
FROM vsds.surveypoints sp
		 JOIN common.systems s ON sp.sysid = s.id
)
SELECT a.*,
       CASE
           WHEN a.corrected_n >= 50
               THEN 50.0 /
                    ((4.0*pi()/3.0) * power(a.maxdistance, 3))
           ELSE a.corrected_n::float /
                ((4.0*pi()/3.0) * power(20.0, 3))
       END AS rho
FROM adjusted a
;
GRANT SELECT ON vsds.v_surveypoints TO edservice;
GRANT SELECT ON vsds.v_surveypoints TO edviewer;

CREATE OR REPLACE VIEW vsds.v_surveys AS
WITH stats AS (
SELECT sp.surveyid,
       max(sp.rho) AS rho_max,
       avg(sp.x) AS x,
       avg(sp.y) AS y,
       stddev_samp(sp.rho) AS rho_stddev,
       jsonb_agg(jsonb_build_object('z', sp.zsample, 'rho', sp.rho)) AS points
FROM vsds.v_surveypoints sp
GROUP BY sp.surveyid
)
SELECT cmdr.name AS cmdrname,
       c.name AS projectname,
       s.*,
       sp.*
FROM vsds.surveys s
     JOIN stats sp ON s.id = sp.surveyid
     JOIN vsds.projects c ON s.projectid = c.id
     JOIN common.cmdrs cmdr ON s.cmdrid = cmdr.id
;
GRANT SELECT ON vsds.v_surveys TO edservice;
GRANT SELECT ON vsds.v_surveys TO edviewer;

CREATE OR REPLACE VIEW vsds.v_projects AS
WITH zsamples AS (
SELECT projectid,
			 array_agg(zsample ORDER BY zsample ASC) AS zsamples
FROM vsds.project_zsamples
GROUP BY projectid
)
SELECT p.id, p.name,
			 zs.zsamples
FROM vsds.projects p
		 LEFT JOIN zsamples zs ON p.id = zs.projectid
;
GRANT SELECT ON vsds.v_projects TO edservice;
GRANT SELECT ON vsds.v_projects TO edviewer;

CREATE OR REPLACE VIEW vsds.v_folders AS
WITH lastproc AS (
SELECT fp.*
FROM vsds.folders f
		 CROSS JOIN LATERAL (
     			 SELECT id, folderid, receivedat, startedat, finishedat
					 FROM vsds.folder_processing
					 WHERE folderid = f.id
					 ORDER BY receivedat DESC
					 LIMIT 1) fp
), procstatus AS (
SELECT ss.folderid, ssp.procid,
			 count(*) FILTER (WHERE ssp.success IS NULL)   AS inprogress,
			 count(*) FILTER (WHERE ssp.success = true)    AS finished,
			 count(*) FILTER (WHERE ssp.success = false)   AS failed
FROM vsds.sheet_processing ssp
		 JOIN vsds.sheets sh ON ssp.sheetid = sh.id
		 JOIN vsds.spreadsheets ss ON sh.spreadsheetid = ss.id
GROUP BY ss.folderid, ssp.procid
)
SELECT f.id AS folderid, f.name, f.gcpid,
			 lp.receivedat, lp.startedat, lp.finishedat,
			 ps.inprogress, ps.finished, ps.failed
FROM vsds.folders f
		 LEFT JOIN lastproc lp ON f.id = lp.folderid
		 LEFT JOIN procstatus ps ON lp.id = ps.procid AND lp.folderid = f.id
;
GRANT SELECT ON vsds.v_folders TO edservice;

CREATE OR REPLACE VIEW vsds.v_spreadsheetvariants AS
SELECT sv.id,
       sv.projectid,
       p.name AS projectname,
       sv.name,
       sv.headerrow,
       sv.sysnamecolumn,
       sv.zsamplecolumn,
       sv.systemcountcolumn,
       sv.maxdistancecolumn,
       COALESCE(
           jsonb_agg(
               jsonb_build_object(
                   'id',    vc.id,
                   'col',   vc.col,
                   'row',   vc.row,
                   'value', vc.value
               ) ORDER BY vc.id
           ) FILTER (WHERE vc.id IS NOT NULL),
           '[]'::jsonb
       ) AS checks
FROM vsds.spreadsheetvariants sv
JOIN vsds.projects p ON p.id = sv.projectid
LEFT JOIN vsds.spreadsheetvariant_checks vc
    ON vc.variantid = sv.id
GROUP BY sv.id, sv.projectid, p.name, sv.name,
         sv.headerrow, sv.sysnamecolumn, sv.zsamplecolumn,
         sv.systemcountcolumn, sv.maxdistancecolumn;
GRANT SELECT ON vsds.v_spreadsheetvariants TO edservice;
GRANT SELECT ON vsds.v_spreadsheetvariants TO edviewer;
