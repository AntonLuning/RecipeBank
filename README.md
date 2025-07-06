# RecipeBank
A recipe management system with AI-powered features.

## System Architecture
![System Architecture](docs/system_architecture.png)

## Sequences
![Sequences](docs/sequences.png)


## General Ideas
- Plan your upcoming dishes (could include AI to help based on cost, diet, etc.)
  - Generate grocery lists (could use AI to group items)


# TODO
- Collect metrics with Prometheus (and show in Grafana)
- Collect logs with Loki (and Promtail)
- Handle auth in API and UI

## API
- Let the AI create some defualt tags (e.g., "dinner/lunch/dessert/etc.", "vegan/fish/meat/etc.")

## UI
- Improved visual design (e.g., "Coop recept")
  - support for different devices (mobile, tablet/ipad, desktop)
  - support for dark/light mode
  - spinners for API dependent actions
- Scale recipes by servings
- Sections for ingredients and instructions
- Validation for editing recipes (client side)
- Language support (english, swedish)
