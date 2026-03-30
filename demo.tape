# env-pilot demo recording
# Run: vhs demo.tape

Require env-pilot

Output screenshots/demo.gif
Output screenshots/demo.mp4

Set Width 1200
Set Height 600
Set FontSize 16
Set Padding 20
Set WindowBar Colorful
Set BorderRadius 8
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms
Set CursorBlink false

# Setup — build and enter example dir
Hide
Type "cd examples && rm -f .env"
Enter
Sleep 0.5
Show

# Launch env-pilot
Type "env-pilot"
Enter
Sleep 1s

# Navigate down through keys, pausing to show the detail panel
Sleep 0.5
Down
Sleep 0.5
Down
Sleep 0.5

# Edit JWT_SECRET (sensitive key) — Tab to edit, type value, confirm
Tab
Sleep 0.3
Type "my-super-secret-jwt-2024"
Sleep 0.5
Enter
Sleep 0.5

# Confirm the sensitive value
Enter
Sleep 0.5

# Skip OAUTH_CLIENT_ID
Type "s"
Sleep 0.3

# Skip OAUTH_CLIENT_SECRET
Type "s"
Sleep 0.5

# Navigate to OPENAI_API_KEY
Sleep 0.3

# Edit it — Tab, type, confirm
Tab
Sleep 0.3
Type "sk-proj-abc123def456"
Sleep 0.5
Enter
Sleep 0.5

# Confirm
Enter
Sleep 0.5

# Skip STRIPE_SECRET_KEY
Type "s"
Sleep 0.3

# Accept SENTRY_DSN default
Tab
Sleep 0.3
Enter
Sleep 0.3

# Navigate through App Config — accept defaults
Tab
Sleep 0.3
Enter
Sleep 0.3

# LOG_LEVEL — type a custom value
Tab
Sleep 0.3
Type "info"
Enter
Sleep 0.3

# DEBUG — type false
Tab
Sleep 0.3
Type "false"
Enter
Sleep 0.5

# Scroll back up to show the completed state
Up@200ms 10
Sleep 1s

# Navigate down slowly to show all the set values
Down@300ms 10
Sleep 1s

# Write & exit
Type "w"
Sleep 1s
