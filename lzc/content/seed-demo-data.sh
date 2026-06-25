#!/bin/sh
COOKIE=/tmp/wt.jar
BASE=http://127.0.0.1:8080
curl -sS -c "$COOKIE" -b "$COOKIE" -X POST "$BASE/user/signin" -H "Content-Type: application/x-www-form-urlencoded" -d "username=admin" -d "password=admin" -o /dev/null
curl -sS -c "$COOKIE" -b "$COOKIE" -X POST "$BASE/self/profile" -H "Content-Type: application/x-www-form-urlencoded" -d "language=zh-Hans" -d "theme=browser" -d "totals_show=running" -d "timezone=Asia/Shanghai" -o /dev/null
for workout in /lzcapp/pkg/content/demo-data/*.gpx /lzcapp/pkg/content/demo-data/*.tcx /lzcapp/pkg/content/demo-data/*.zip; do
  [ -f "$workout" ] || continue
  curl -sS -c "$COOKIE" -b "$COOKIE" -X POST "$BASE/workouts" -F "file=@${workout}" -F "type=auto" -F "notes=LazyCat demo import: $(basename "$workout")" -o /dev/null
done
curl -sS -c "$COOKIE" -b "$COOKIE" -X POST "$BASE/daily" -d "date=2026-06-20" -d "weight=71.8" -d "weight_unit=kg" -d "steps=8500" -o /dev/null
curl -sS -c "$COOKIE" -b "$COOKIE" -X POST "$BASE/daily" -d "date=2026-06-21" -d "weight=71.6" -d "weight_unit=kg" -d "steps=9200" -o /dev/null
curl -sS -c "$COOKIE" -b "$COOKIE" -X POST "$BASE/daily" -d "date=2026-06-22" -d "weight=71.5" -d "weight_unit=kg" -d "steps=10300" -o /dev/null
curl -sS -c "$COOKIE" -b "$COOKIE" -X POST "$BASE/daily" -d "date=2026-06-23" -d "weight=71.2" -d "weight_unit=kg" -d "steps=7800" -o /dev/null
curl -sS -c "$COOKIE" -b "$COOKIE" -X POST "$BASE/daily" -d "date=2026-06-24" -d "weight=71.0" -d "weight_unit=kg" -d "steps=11200" -o /dev/null
curl -sS -c "$COOKIE" -b "$COOKIE" -X POST "$BASE/daily" -d "date=2026-06-25" -d "weight=70.9" -d "weight_unit=kg" -d "steps=13500" -o /dev/null
echo seed-ok
