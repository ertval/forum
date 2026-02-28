# Home - User Dropdown and Notification Badge

URL: http://localhost:8080/
Title: Home - Forum

## Snapshot excerpt
- User menu expanded with actions:
  - `🔔 My Activity`
  - `Settings`
  - `Logout`
- Notification badge element present in action link:
  - class: `notification-badge`
  - text: `0`
  - hidden: `true`

## DOM assertion capture
```json
{
  "actions": [
    {"text": "🔔 My Activity 0", "href": "/activity"},
    {"text": "Settings", "href": "/settings"},
    {"text": "Logout", "href": "/logout"}
  ],
  "badge": {"text": "0", "hidden": true, "className": "notification-badge"}
}
```
