#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DB_PATH="${DATABASE_PATH:-${PROJECT_ROOT}/data/forum.db}"
UPLOADS_DIR="${PROJECT_ROOT}/static/uploads"
MODE="dry-run"

usage() {
    cat <<EOF
Usage: bash scripts/audit/clean_dead_refs.sh [--dry-run|--apply] [--db <path>]

Checks and optionally cleans dead references:
  - Template route refs that do not match registered routes
  - Template static asset refs that do not exist on disk
  - posts.image_path entries pointing to missing files in static/uploads
  - Orphaned reactions / notifications / reports (when confidently orphaned)

Modes:
  --dry-run   Report only (default)
  --apply     Apply DB cleanup actions

Options:
  --db <path> Override sqlite DB path
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run)
            MODE="dry-run"
            shift
            ;;
        --apply)
            MODE="apply"
            shift
            ;;
        --db)
            DB_PATH="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown arg: $1"
            usage
            exit 1
            ;;
    esac
done

if ! command -v sqlite3 >/dev/null 2>&1; then
    echo "ERROR: sqlite3 is required."
    exit 1
fi

if [[ ! -f "$DB_PATH" ]]; then
    echo "ERROR: DB not found at: $DB_PATH"
    exit 1
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

sql() {
    sqlite3 -noheader -separator '|' "$DB_PATH" "$1"
}

table_exists() {
    local t="$1"
    [[ "$(sql "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='${t}';")" -gt 0 ]]
}

normalize_image_path() {
    local p="$1"
    p="${p#./}"
    p="${p#/}"
    p="${p#/static/}"
    p="${p#static/}"
    p="${p#uploads/}"
    p="${p#static/uploads/}"
    p="${p#uploads/}"
    printf '%s' "$p"
}

extract_route_patterns() {
    local out="$1"
    grep -RhoE 'Handle(Func)?\("[A-Z]+ /[^\"]*' "${PROJECT_ROOT}/internal" "${PROJECT_ROOT}/cmd" \
        | sed -E 's/^.*\("[A-Z]+ (\/[^\"]*)$/\1/' \
        | sort -u > "$out"
}

route_ref_is_valid() {
    local path="$1"
    local route
    while IFS= read -r route; do
        [[ -z "$route" ]] && continue

        if [[ "$route" == "/static/" ]]; then
            [[ "$path" == /static/* ]] && return 0
            continue
        fi

        local route_regex
        route_regex="$(printf '%s' "$route" | sed -E 's#\{[^/]+\}#[^/]+#g')"
        if [[ "$path" =~ ^${route_regex}$ ]]; then
            return 0
        fi
    done < "$ROUTE_PATTERNS_FILE"

    return 1
}

extract_template_refs() {
    local out="$1"
    grep -RhoE '(href|src|action)="[^"]+"' "${PROJECT_ROOT}/templates"/*.html \
        | sed -E 's/^[a-z]+="//; s/"$//' \
        | grep '^/' \
        | sed -E 's/[?#].*$//' \
        | sort -u > "$out"
}

echo "=== Dead Reference Audit (${MODE}) ==="
echo "DB: $DB_PATH"
echo

ROUTE_PATTERNS_FILE="$TMP_DIR/route_patterns.txt"
TEMPLATE_REFS_FILE="$TMP_DIR/template_refs.txt"
DEAD_ROUTE_REFS_FILE="$TMP_DIR/dead_route_refs.txt"
DEAD_ASSET_REFS_FILE="$TMP_DIR/dead_asset_refs.txt"
MISSING_IMAGE_ROWS_FILE="$TMP_DIR/missing_image_rows.txt"
REACTION_ORPHANS_FILE="$TMP_DIR/reaction_orphans.txt"
NOTIFICATION_ORPHANS_FILE="$TMP_DIR/notification_orphans.txt"
REPORT_ORPHANS_FILE="$TMP_DIR/report_orphans.txt"

extract_route_patterns "$ROUTE_PATTERNS_FILE"
extract_template_refs "$TEMPLATE_REFS_FILE"

> "$DEAD_ROUTE_REFS_FILE"
> "$DEAD_ASSET_REFS_FILE"

while IFS= read -r ref; do
    [[ -z "$ref" ]] && continue
    if [[ "$ref" == /static/* ]]; then
        local_path="${PROJECT_ROOT}${ref}"
        if [[ ! -f "$local_path" ]]; then
            echo "$ref" >> "$DEAD_ASSET_REFS_FILE"
        fi
        continue
    fi

    check_path="$(printf '%s' "$ref" | sed -E 's/\{\{[^}]+\}\}/x/g')"
    if ! route_ref_is_valid "$check_path"; then
        echo "$ref" >> "$DEAD_ROUTE_REFS_FILE"
    fi
done < "$TEMPLATE_REFS_FILE"

sort -u "$DEAD_ROUTE_REFS_FILE" -o "$DEAD_ROUTE_REFS_FILE"
sort -u "$DEAD_ASSET_REFS_FILE" -o "$DEAD_ASSET_REFS_FILE"

dead_route_count="$(wc -l < "$DEAD_ROUTE_REFS_FILE" | tr -d ' ')"
dead_asset_count="$(wc -l < "$DEAD_ASSET_REFS_FILE" | tr -d ' ')"

echo "[Template Route Refs] dead=${dead_route_count}"
if [[ "$dead_route_count" -gt 0 ]]; then
    cat "$DEAD_ROUTE_REFS_FILE"
fi
echo

echo "[Template Static Asset Refs] missing=${dead_asset_count}"
if [[ "$dead_asset_count" -gt 0 ]]; then
    cat "$DEAD_ASSET_REFS_FILE"
fi
echo

> "$MISSING_IMAGE_ROWS_FILE"
if table_exists posts; then
    while IFS='|' read -r id public_id image_path; do
        [[ -z "$id" ]] && continue
        normalized="$(normalize_image_path "$image_path")"
        if [[ -z "$normalized" || "$normalized" == *".."* ]]; then
            echo "${id}|${public_id}|${image_path}|${normalized}" >> "$MISSING_IMAGE_ROWS_FILE"
            continue
        fi

        if [[ ! -f "${UPLOADS_DIR}/${normalized}" ]]; then
            echo "${id}|${public_id}|${image_path}|${normalized}" >> "$MISSING_IMAGE_ROWS_FILE"
        fi
    done < <(sql "SELECT id, public_id, image_path FROM posts WHERE image_path IS NOT NULL AND TRIM(image_path) <> '';")
fi

missing_image_count="$(wc -l < "$MISSING_IMAGE_ROWS_FILE" | tr -d ' ')"
echo "[posts.image_path] missing_files=${missing_image_count}"
if [[ "$missing_image_count" -gt 0 ]]; then
    awk -F'|' '{print "- post_public_id=" $2 ", image_path=" $3 ", normalized=" $4}' "$MISSING_IMAGE_ROWS_FILE"
fi
echo

> "$REACTION_ORPHANS_FILE"
if table_exists reactions; then
    sql "
        SELECT r.id, r.public_id, 'missing_post', r.target_type, r.target_id
        FROM reactions r
        LEFT JOIN posts p ON r.target_type='post' AND r.target_id=p.id
        WHERE r.target_type='post' AND p.id IS NULL
        UNION ALL
        SELECT r.id, r.public_id, 'missing_comment', r.target_type, r.target_id
        FROM reactions r
        LEFT JOIN comments c ON r.target_type='comment' AND r.target_id=c.id
        WHERE r.target_type='comment' AND c.id IS NULL
        UNION ALL
        SELECT r.id, r.public_id, 'invalid_target_type', r.target_type, r.target_id
        FROM reactions r
        WHERE r.target_type NOT IN ('post', 'comment')
        UNION ALL
        SELECT r.id, r.public_id, 'missing_user', r.target_type, r.target_id
        FROM reactions r
        LEFT JOIN users u ON r.user_id=u.id
        WHERE u.id IS NULL;
    " > "$REACTION_ORPHANS_FILE"
fi

reaction_orphan_count="$(wc -l < "$REACTION_ORPHANS_FILE" | tr -d ' ')"
echo "[reactions] orphaned=${reaction_orphan_count}"
if [[ "$reaction_orphan_count" -gt 0 ]]; then
    awk -F'|' '{print "- reaction_public_id=" $2 ", reason=" $3 ", target_type=" $4 ", target_id=" $5}' "$REACTION_ORPHANS_FILE"
fi
echo

> "$NOTIFICATION_ORPHANS_FILE"
if table_exists notifications; then
    sql "
        SELECT n.id, n.public_id, 'missing_target_post', n.target_id, n.type
        FROM notifications n
        LEFT JOIN posts p ON n.target_id=p.id
        WHERE p.id IS NULL
        UNION ALL
        SELECT n.id, n.public_id, 'missing_recipient_user', n.target_id, n.type
        FROM notifications n
        LEFT JOIN users u ON n.user_id=u.id
        WHERE u.id IS NULL
        UNION ALL
        SELECT n.id, n.public_id, 'missing_actor_user', n.target_id, n.type
        FROM notifications n
        LEFT JOIN users u ON n.actor_id=u.id
        WHERE u.id IS NULL;
    " > "$NOTIFICATION_ORPHANS_FILE"
fi

notification_orphan_count="$(wc -l < "$NOTIFICATION_ORPHANS_FILE" | tr -d ' ')"
echo "[notifications] orphaned=${notification_orphan_count}"
if [[ "$notification_orphan_count" -gt 0 ]]; then
    awk -F'|' '{print "- notification_public_id=" $2 ", reason=" $3 ", target_id=" $4 ", type=" $5}' "$NOTIFICATION_ORPHANS_FILE"
fi
echo

> "$REPORT_ORPHANS_FILE"
report_table_present=0
if table_exists reports; then
    report_table_present=1
    sql "
        SELECT r.id, r.public_id, 'missing_post', r.target_type, r.target_id
        FROM reports r
        LEFT JOIN posts p ON r.target_type='post' AND r.target_id=p.id
        WHERE r.target_type='post' AND p.id IS NULL
        UNION ALL
        SELECT r.id, r.public_id, 'missing_comment', r.target_type, r.target_id
        FROM reports r
        LEFT JOIN comments c ON r.target_type='comment' AND r.target_id=c.id
        WHERE r.target_type='comment' AND c.id IS NULL;
    " > "$REPORT_ORPHANS_FILE"
fi

report_orphan_count="$(wc -l < "$REPORT_ORPHANS_FILE" | tr -d ' ')"
if [[ "$report_table_present" -eq 1 ]]; then
    echo "[reports] orphaned=${report_orphan_count}"
    if [[ "$report_orphan_count" -gt 0 ]]; then
        awk -F'|' '{print "- report_public_id=" $2 ", reason=" $3 ", target_type=" $4 ", target_id=" $5}' "$REPORT_ORPHANS_FILE"
    fi
    echo
else
    echo "[reports] table not present, skipped"
    echo
fi

cleaned_missing_images=0
cleaned_reactions=0
cleaned_notifications=0
cleaned_reports=0

if [[ "$MODE" == "apply" ]]; then
    if [[ "$missing_image_count" -gt 0 ]]; then
        image_ids_csv="$(awk -F'|' '{print $1}' "$MISSING_IMAGE_ROWS_FILE" | paste -sd, -)"
        sql "UPDATE posts SET image_path = NULL, updated_at = CURRENT_TIMESTAMP WHERE id IN (${image_ids_csv});"
        cleaned_missing_images="$missing_image_count"
    fi

    if [[ "$reaction_orphan_count" -gt 0 ]]; then
        reaction_ids_csv="$(awk -F'|' '{print $1}' "$REACTION_ORPHANS_FILE" | sort -u | paste -sd, -)"
        reaction_unique_count="$(awk -F'|' '{print $1}' "$REACTION_ORPHANS_FILE" | sort -u | wc -l | tr -d ' ')"
        sql "DELETE FROM reactions WHERE id IN (${reaction_ids_csv});"
        cleaned_reactions="$reaction_unique_count"
    fi

    if [[ "$notification_orphan_count" -gt 0 ]]; then
        notification_ids_csv="$(awk -F'|' '{print $1}' "$NOTIFICATION_ORPHANS_FILE" | sort -u | paste -sd, -)"
        notification_unique_count="$(awk -F'|' '{print $1}' "$NOTIFICATION_ORPHANS_FILE" | sort -u | wc -l | tr -d ' ')"
        sql "DELETE FROM notifications WHERE id IN (${notification_ids_csv});"
        cleaned_notifications="$notification_unique_count"
    fi

    if [[ "$report_orphan_count" -gt 0 ]]; then
        report_ids_csv="$(awk -F'|' '{print $1}' "$REPORT_ORPHANS_FILE" | sort -u | paste -sd, -)"
        report_unique_count="$(awk -F'|' '{print $1}' "$REPORT_ORPHANS_FILE" | sort -u | wc -l | tr -d ' ')"
        sql "DELETE FROM reports WHERE id IN (${report_ids_csv});"
        cleaned_reports="$report_unique_count"
    fi
fi

echo "=== Summary ==="
echo "mode=${MODE}"
echo "template_dead_route_refs=${dead_route_count}"
echo "template_missing_static_assets=${dead_asset_count}"
echo "posts_missing_image_refs=${missing_image_count}"
echo "reaction_orphans=${reaction_orphan_count}"
echo "notification_orphans=${notification_orphan_count}"
if [[ "$report_table_present" -eq 1 ]]; then
    echo "report_orphans=${report_orphan_count}"
fi

if [[ "$MODE" == "apply" ]]; then
    echo "cleaned_posts_image_path=${cleaned_missing_images}"
    echo "cleaned_reactions=${cleaned_reactions}"
    echo "cleaned_notifications=${cleaned_notifications}"
    if [[ "$report_table_present" -eq 1 ]]; then
        echo "cleaned_reports=${cleaned_reports}"
    fi
fi
