# nvolt Website

This directory contains the landing page and documentation for nvolt.

## Files

- `index.html` - Landing page
- `docs.html` - Documentation
- `styles.css` - Main stylesheet
- `docs.css` - Documentation-specific styles
- `script.js` - Interactive functionality
- `assets/logo.png` - nvolt logo

## Icons

The website uses [Lucide Icons](https://lucide.dev) - a free, open-source icon library loaded via CDN.

## Development

To preview the website locally:

```bash
cd web
python3 -m http.server 8000
```

Then visit [http://localhost:8000](http://localhost:8000)

## Deployment

### GitHub Pages

1. Push the `web/` directory to your repository
2. Go to Settings > Pages
3. Set source to the branch with your web files
4. Set folder to `/web`

### Netlify

```bash
# Deploy directly
cd web
npx netlify-cli deploy --prod
```

### Vercel

```bash
cd web
npx vercel --prod
```

## Customization

### Colors

The color scheme is defined in `styles.css` using CSS variables:

```css
--primary-yellow: #FFC107;
--primary-black: #1a1a1a;
--dark-bg: #0a0a0a;
```

### Content

- Landing page copy: Edit `index.html`
- Documentation: Edit `docs.html`
- Styles: Edit `styles.css` and `docs.css`

## SEO

The landing page includes:
- Meta descriptions and keywords
- Open Graph tags for social media
- Schema.org structured data
- Semantic HTML5 markup

## Browser Support

- Modern browsers (Chrome, Firefox, Safari, Edge)
- Mobile responsive design
- Progressive enhancement approach
