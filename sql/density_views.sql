CREATE OR REPLACE VIEW density.v_surveypoints AS
WITH adjusted AS (
SELECT sp.id, sp.surveyid, s.name AS sysname,
       sp.zsample, s.x, s.y, s.z,
       greatest(least(sp.syscount, 50), 1) AS syscount,
       greatest(least(sp.maxdistance, 20), 1) AS maxdistance
FROM density.surveypoints sp
		 JOIN common.systems s ON sp.sysid = s.id
)
SELECT a.*,
       a.syscount/((4*pi()/3)*power(a.maxdistance, 3)) AS rho
FROM adjusted a
;
GRANT SELECT ON density.v_surveypoints TO edservice;
GRANT SELECT ON density.v_surveypoints TO edviewer;

CREATE OR REPLACE VIEW density.v_surveys AS
WITH stats AS (
SELECT sp.surveyid,
       max(sp.rho) AS rho_max,
       avg(sp.x) AS x,
       avg(sp.y) AS y,
       stddev_samp(sp.rho) AS rho_stddev,
       jsonb_agg(jsonb_build_object('z', sp.zsample, 'rho', sp.rho)) AS points
FROM density.v_surveypoints sp
GROUP BY sp.surveyid
)
SELECT cmdr.name AS cmdrname,
       c.name AS campaignname,
       s.*,
       sp.*
FROM density.surveys s
     JOIN stats sp ON s.id = sp.surveyid
     JOIN density.campaigns c ON s.campaignid = c.id
     JOIN common.cmdrs cmdr ON s.cmdrid = cmdr.id
;
GRANT SELECT ON density.v_surveys TO edservice;
GRANT SELECT ON density.v_surveys TO edviewer;
