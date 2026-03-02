# Activity Page Layout - After Patch

URL: http://localhost:8080/activity
Title: My Activity - Forum

## Snapshot excerpt
- Left filter card visible with title `Filter Activity`.
- Section headers are clickable links:
  - `Created Posts` -> `/board?my_posts=true`
  - `Post Reactions` -> `/board?liked_posts=true`
  - `Comments` -> `/comments`
- Reaction symbols rendered and right-aligned in each reaction row.

## DOM assertion capture
```json
{
  "headers": [
    {"text": "Created Posts", "href": "/board?my_posts=true"},
    {"text": "Post Reactions", "href": "/board?liked_posts=true"},
    {"text": "Comments", "href": "/comments"}
  ],
  "symbols": [
    {"text": "👍", "color": "rgb(140, 201, 181)", "className": "activity-reaction-symbol activity-reaction-symbol-like"},
    {"text": "👎", "color": "rgb(211, 47, 47)", "className": "activity-reaction-symbol activity-reaction-symbol-dislike"}
  ]
}
```
