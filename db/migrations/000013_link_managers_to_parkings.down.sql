UPDATE parkings
SET manager_id = NULL
WHERE manager_id IS NOT NULL;