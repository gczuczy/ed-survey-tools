
CREATE OR REPLACE FUNCTION density.addsheetsurvey(cmdr text, campaign text) RETURNS int AS $$
DECLARE
	cmdrid int;
	campaignid int;
	mid int;
BEGIN
   SELECT INTO cmdrid id FROM common.cmdrs WHERE name = cmdr;
   IF NOT FOUND THEN
      INSERT INTO common.cmdrs (name) VALUES (cmdr) RETURNING id INTO cmdrid;
   END IF;

   SELECT INTO campaignid id FROM density.campaigns WHERE name = campaign;
   IF NOT FOUND THEN
      INSERT INTO density.campaigns (name) VALUES (campaign)
      RETURNING id INTO campaignid;
   END IF;

   INSERT INTO density.surveys (cmdrid, campaignid) VALUES (cmdrid, campaignid)
   RETURNING id INTO mid;

   RETURN mid;
END;
$$ LANGUAGE plpgsql VOLATILE STRICT PARALLEL UNSAFE SECURITY INVOKER;

GRANT EXECUTE ON FUNCTION density.addsheetsurvey(cmdr text, campaign text) TO edservice;
