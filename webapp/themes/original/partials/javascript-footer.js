{{ define "javascript-footer" }}

// run when full page body has loaded
document.addEventListener( "DOMContentLoaded", () => {
});

// Log metadata to console
console.log( "Welcome to EventHunt!\n  version: {{ .App.version }}\n  environment: {{ .App.environment }}" );

{{ end }}
