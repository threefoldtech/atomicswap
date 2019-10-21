package timings

import "time"

//LockTime is the default time an atomic swap is locked before a refund can be issued
const LockTime = 48 * time.Hour
