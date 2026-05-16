Output demo.gif

Set Shell "zsh"

Set FontSize 16
Set Width 1200
Set Height 700

# Optional pacing
Set TypingSpeed 40ms

# Start shell
Type "source ~/.zshrc >/dev/null 2>&1"
Enter

# Wait for prompt to fully settle
Sleep 5000ms

# Clear setup artifacts completely
Type "clear"
Enter

Sleep 3000ms

# Now begin visible demo
Type "shai 'Show me a pretty git log graph with one line per commit'"
Enter

Sleep 6s
Enter

Sleep 3s