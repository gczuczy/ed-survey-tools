
CREATE OR REPLACE FUNCTION common.logincmdr(cmdr text, cid bigint) RETURNS common.cmdrs AS $$
DECLARE
	ret record;
BEGIN
   SELECT INTO ret * FROM common.cmdrs WHERE customerid = cid;
   IF FOUND THEN
	 		RETURN ret;
   END IF;

	 SELECT INTO ret * FROM common.cmdrs WHERE name = cmdr;
   IF FOUND THEN
	 		RETURN ret;
   END IF;

	 INSERT INTO common.cmdrs (name, customerid) VALUES (cmdr, cid)
	 RETURNING * INTO ret;
	 RETURN ret;
END;
$$ LANGUAGE plpgsql VOLATILE STRICT PARALLEL UNSAFE SECURITY INVOKER;

GRANT EXECUTE ON FUNCTION density.addsheetsurvey(cmdr text, campaign text) TO edservice;
