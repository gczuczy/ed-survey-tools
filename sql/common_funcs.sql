
CREATE OR REPLACE FUNCTION common.logincmdr(cmdr text, cid bigint) RETURNS common.v_cmdrs AS $$
DECLARE
	cmdrid int;
	ret record;
BEGIN
   SELECT INTO ret * FROM common.v_cmdrs WHERE customerid = cid;
   IF FOUND THEN
	 		RETURN ret;
   END IF;

	 SELECT INTO ret * FROM common.v_cmdrs WHERE name = cmdr;
   IF FOUND THEN
	 		RETURN ret;
   END IF;

	 INSERT INTO common.cmdrs (name, customerid) VALUES (cmdr, cid)
	 RETURNING id INTO cmdrid;
	 SELECT INTO ret * FROM common.v_cmdrs WHERE id = cmdrid;
	 RETURN ret;
END;
$$ LANGUAGE plpgsql VOLATILE STRICT PARALLEL UNSAFE SECURITY INVOKER;

GRANT EXECUTE ON FUNCTION common.logincmdr(cmdr text, cid bigint) TO edservice;
