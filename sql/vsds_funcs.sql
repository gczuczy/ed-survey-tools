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
