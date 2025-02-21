import js from "@eslint/js";
import globals from "globals";

export default [
  {
    name: "calendar/recommended-rules",
    files: ["html/**/*.js"],
    rules: js.configs.recommended.rules,
    languageOptions: {
      sourceType: "script",
      globals: {
        ...globals.browser,
      },
    },
  },
];
