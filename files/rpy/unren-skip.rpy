# UnRen Skip Unseen Content Enabler
# This file allows skipping all text including unseen content
# Skip controls: TAB (hold) or CTRL (toggle)

init 999 python:
    _preferences.skip_unseen = True
    renpy.game.preferences.skip_unseen = True
    renpy.config.allow_skipping = True
    renpy.config.fast_skipping = True
