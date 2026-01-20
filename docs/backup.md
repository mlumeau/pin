# Backup and Restore

## Create a backup
```bash
go run . backup
```

Optional destination:
```bash
go run . backup /path/to/pin-backup.zip
```

The backup includes:
- `identity.db`
- `uploads/` (from `PIN_UPLOADS_DIR`)

## Restore
1. Stop the server.
2. Unzip the backup.
3. Copy `identity.db` to your configured `PIN_DB_PATH`.
4. Copy the `uploads/` folder to your configured `PIN_UPLOADS_DIR`.
5. Start the server.

Tip: keep file ownership consistent with the service user.
