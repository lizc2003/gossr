module.exports = {
    "plugins": [
        ["@babel/plugin-transform-runtime", {
            'corejs': 3
        }]
    ],
    "presets": [
        [
            "@babel/preset-env"
        ]
    ],
    "ignore": ["node_modules/*"]
}
