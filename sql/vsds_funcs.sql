-- Survey insertion is handled by the Go processing layer via
-- direct SQL statements within the long batch transaction.

-- Refreshes both survey materialized views in dependency order.
-- Defined with SECURITY DEFINER so edservice can invoke it even
-- though the matviews are owned by edadmin.
CREATE OR REPLACE FUNCTION vsds.refresh_survey_matviews()
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
    REFRESH MATERIALIZED VIEW vsds.v_surveypoints;
    REFRESH MATERIALIZED VIEW vsds.v_surveys;
END;
$$;
GRANT EXECUTE ON FUNCTION vsds.refresh_survey_matviews()
    TO edservice;

-- Aggregate survey point density into spatial voxels.
-- xz_step: cell size on the galactic X-Z plane (ly)
-- y_step:  cell size on the galactic Y axis (ly)
-- Returns one row per non-empty voxel with min/avg/max rho.
CREATE OR REPLACE FUNCTION vsds.sectors(
    xz_step FLOAT,
    y_step  FLOAT
) RETURNS TABLE (
    gc_x    DOUBLE PRECISION,
    gc_z    DOUBLE PRECISION,
    y_min   DOUBLE PRECISION,
    y_max   DOUBLE PRECISION,
    rho_min DOUBLE PRECISION,
    rho_avg DOUBLE PRECISION,
    rho_max DOUBLE PRECISION,
    cnt     BIGINT
)
LANGUAGE SQL
STABLE
AS $$
  SELECT
    floor(gc_x / xz_step) * xz_step + (xz_step / 2.0),
    floor(gc_z / xz_step) * xz_step + (xz_step / 2.0),
    floor(gc_y / y_step)  * y_step,
    floor(gc_y / y_step)  * y_step + y_step,
    min(rho),
    avg(rho),
    max(rho),
    count(*)
  FROM vsds.v_surveypoints
  GROUP BY 1, 2, 3, 4
  ORDER BY 1, 2, 3
$$;
GRANT EXECUTE ON FUNCTION vsds.sectors(FLOAT, FLOAT)
    TO edservice;
