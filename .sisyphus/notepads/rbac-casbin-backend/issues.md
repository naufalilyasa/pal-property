# Issues

- The plan’s Task 1 acceptance grep for `AutoMigrate` is stricter than the official adapter API naming, because the safe disable call itself contains the string `TurnOffAutoMigrate`. Runtime behavior is correct (no schema auto-creation path is used), but the literal grep line may need interpretation during final review.
