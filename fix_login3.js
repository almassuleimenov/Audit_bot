const fs = require('fs');
const file = 'audit-dashboard/src/app/login/page.tsx';
let data = fs.readFileSync(file, 'utf8');
data = data.replace(
  "const API_URL = process.env.NEXT_PUBLIC_API_URL || 'https://audit-bot-1-j1xj.onrender.com';\n      const res = await fetch(`${API_URL}/api/login`, {",
  "const res = await fetch('/api/login', {"
);
fs.writeFileSync(file, data);
