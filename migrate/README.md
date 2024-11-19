# Operations

-   Create a new migration with:

```sh
make migration create_some_table
```

This will create both up and down files

-   Run all migrations:

```sh
make migrate-up
```

-   If some migration got corrupted:

```sh
# Suppose migration 15 is corrupted, you can go back to
# previous migration 14 with
make force-migration 14
```
