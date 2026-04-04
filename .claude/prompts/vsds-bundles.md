We're designing the first part of a series of features now. Ideas need to be validated as part of the design process.

We have VSDS scans, data is collected. This part is about creating static bundles of these data, that later can be used for visual presentation without hammering the DB.

How I envision it:
 - a bundle can include all, one, or more projects' data (all must be an implicit all, without explicit listing)
 - Bundles are saved at a predefined/configured location
 - Bundles have predefined names, something like `$measurementType-$hash.json.gz`, az the measurementtype would be vsds for vsds
 - They are JSON data, preemptively gzip compressed - Verification needed whether UI can load it as plain json this way
 - Two serving options has to be supported:
   - For development, serve it from the backend from something like `/static/`
   - For production, it would be served from a `static.` subdomain
 - Bundles would get their part in the VSDS section of the UI, being able to define+configure them, and re-generate (or set to auto-regen after processing)
 - VSDS would have 2 types of bundles: surveypoints(raw surveypoints) and surveys(surveypoints organized into by the "columns" they are taken) - this "subtype" specific to vsds has to be taken into account
 - Bundles would have DB presence, open for later adding different kind of measurement types as well. Separate db schema(namespace) to be considered

# Notes

If we do autoregen, that must be spawned in a separate goroutine and out of the processing txn
