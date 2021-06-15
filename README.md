# filestore-cleanup

A simple utility which will run `filestore/verify` then remove every single block that points to a file that no longer exists, even if pinned(!).
