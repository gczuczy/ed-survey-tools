CREATE OR REPLACE VIEW common.v_cmdrs AS
SELECT c.id, c.name, c.customerid,
			 c.isowner,
			 (c.isowner OR c.isadmin) AS isadmin
FROM common.cmdrs c
;
GRANT SELECT ON common.v_cmdrs TO edservice;
GRANT SELECT ON common.v_cmdrs TO edviewer;
