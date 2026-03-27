CREATE OR REPLACE VIEW bundles.v_bundles AS
SELECT id, measurementtype, name, filename,
       generatedat, autoregen, status, errormessage
FROM bundles.bundles;
GRANT SELECT ON bundles.v_bundles TO edservice;
GRANT SELECT ON bundles.v_bundles TO edviewer;

CREATE OR REPLACE VIEW bundles.v_vsds_bundles AS
SELECT b.id, b.measurementtype, b.name, b.filename,
       b.generatedat, b.autoregen,
       b.status, b.errormessage,
       vb.subtype, vb.allprojects,
       CASE
           WHEN vb.allprojects THEN NULL
           ELSE array_agg(
               p.name ORDER BY p.name
           ) FILTER (WHERE p.name IS NOT NULL)
       END AS projects
FROM bundles.bundles b
JOIN bundles.vsds_bundles vb ON vb.bundleid = b.id
LEFT JOIN bundles.vsds_bundle_projects vbp
    ON vbp.bundleid = b.id
LEFT JOIN vsds.projects p ON p.id = vbp.projectid
WHERE b.measurementtype = 'vsds'
GROUP BY b.id, b.measurementtype, b.name, b.filename,
         b.generatedat, b.autoregen, b.status,
         b.errormessage, vb.subtype, vb.allprojects;
GRANT SELECT ON bundles.v_vsds_bundles TO edservice;
GRANT SELECT ON bundles.v_vsds_bundles TO edviewer;
