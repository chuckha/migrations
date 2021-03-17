# Migrations

Very simple migrations.

Create a directory called "migrations"

Create a migration named `1.sql` and then `2.sql` or whatever number-based naming pattern you want.

Format a migration like this:

```sql
-- up
CREATE TABLE test;
-- SPLIT --
-- down
DROP TABLE test;
```

Note: down is not yet supported.
