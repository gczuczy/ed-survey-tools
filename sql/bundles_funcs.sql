CREATE OR REPLACE FUNCTION bundles.create_vsds_bundle(
    p_name        varchar,
    p_autoregen   bool,
    p_subtype     varchar,
    p_allprojects bool,
    p_projects    int[]
) RETURNS SETOF bundles.v_vsds_bundles AS $$
DECLARE
    v_id int;
BEGIN
    INSERT INTO bundles.bundles
        (measurementtype, name, filename, autoregen)
    VALUES
        ('vsds', p_name, '', p_autoregen)
    RETURNING id INTO v_id;

    UPDATE bundles.bundles
    SET filename = 'vsds-' || v_id || '.json.gz'
    WHERE id = v_id;

    INSERT INTO bundles.vsds_bundles
        (bundleid, subtype, allprojects)
    VALUES
        (v_id, p_subtype, p_allprojects);

    IF NOT p_allprojects THEN
        INSERT INTO bundles.vsds_bundle_projects
            (bundleid, projectid)
        SELECT v_id, unnest(p_projects);
    END IF;

    RETURN QUERY
        SELECT * FROM bundles.v_vsds_bundles
        WHERE id = v_id;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION bundles.queue_autoregen_bundles(
    p_project_ids int[]
) RETURNS void AS $$
BEGIN
    UPDATE bundles.bundles b
    SET status = 'queued'
    FROM bundles.vsds_bundles vb
    LEFT JOIN bundles.vsds_bundle_projects vbp
        ON vbp.bundleid = vb.bundleid
    WHERE b.id = vb.bundleid
      AND b.measurementtype = 'vsds'
      AND b.autoregen = true
      AND b.status != 'generating'
      AND (
          vb.allprojects = true
          OR vbp.projectid = ANY(p_project_ids)
      );
END;
$$ LANGUAGE plpgsql;
