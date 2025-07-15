# FinSights AI Frontend

A modern React-based frontend for the FinSights AI stock screening platform, built with React Router v7, TypeScript, and Tailwind CSS.

## Features

- **Modern Stock Screener Interface** - Advanced filtering with real-time results
- **Server-Side Rendering** - Fast initial page loads with React Router
- **Responsive Design** - Mobile-first design with Tailwind CSS
- **Advanced Filtering** - Multiple filter types with preset strategies
- **Real-time Updates** - Dynamic filtering and sorting
- **Value Investment Focus** - Specialized metrics for value investors

## Tech Stack

- **React Router v7** - Full-stack React framework with SSR
- **TypeScript** - Type-safe development
- **Tailwind CSS v4** - Utility-first CSS framework
- **Vite** - Fast build tool and dev server

## Getting Started

### Prerequisites

- Node.js 18+ 
- npm or yarn
- Backend API running on port 8080 (see backend documentation)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd finsights-ai/frontend
   ```

2. **Install dependencies**
   ```bash
   npm install
   ```

3. **Start development server**
   ```bash
   npm run dev
   ```

4. **Open your browser**
   ```
   http://localhost:5173
   ```

### Build for Production

```bash
npm run build
npm run start
```

## Project Structure

```
frontend/
├── app/
│   ├── routes/
│   │   ├── home.tsx           # Landing page
│   │   └── screener.tsx       # Stock screener interface
│   ├── app.css               # Global styles
│   ├── root.tsx              # Root layout
│   └── routes.ts             # Route configuration
├── public/                   # Static assets
├── package.json             # Dependencies and scripts
└── vite.config.ts          # Vite configuration
```

## Features Overview

### Stock Screener (`/screener`)

- **Quick Filters**: Pre-configured strategies (Value, Dividend, Growth, etc.)
- **Advanced Filters**: Custom filter builder with multiple conditions
- **Sorting Options**: Sort by P/E, ROE, dividend yield, margin of safety, etc.
- **Pagination**: Configurable results per page
- **Responsive Table**: Mobile-optimized data display

### Home Page (`/`)

- **Hero Section**: Platform introduction
- **Feature Highlights**: Key capabilities overview
- **Call-to-Action**: Direct navigation to screener

## API Integration

The frontend connects to the backend API at `http://localhost:8080/api/screener`.

### Filter Format

Filters are sent as JSON arrays in the EODHD-compatible format:
```json
[["pe_ratio","<",15],["roe",">",0.15]]
```

### Supported Filter Fields

- `pe_ratio` - Price-to-earnings ratio
- `roe` - Return on equity
- `dividend_yield` - Dividend yield percentage
- `dividend_growth_5y` - 5-year dividend growth rate
- `margin_of_safety` - Margin of safety percentage
- `earnings_outlook` - Earnings outlook (positive, negative, neutral, stable)
- `close` - Current stock price
- `price_vs_sma50` - Price relative to 50-day SMA
- `price_vs_sma200` - Price relative to 200-day SMA

## Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run start` - Start production server
- `npm run typecheck` - Run TypeScript type checking

## Environment Configuration

The frontend automatically connects to the backend at `http://localhost:8080`. To change this:

1. Update the API URL in `app/routes/screener.tsx` loader function
2. Modify the fetch URL to point to your backend instance

## Styling

The project uses Tailwind CSS v4 with a custom configuration:

- **Font**: Inter font family
- **Dark Mode**: Automatic based on system preference
- **Colors**: Blue-based color scheme for primary actions
- **Components**: Utility-first approach with custom component classes

## Development Guidelines

### Adding New Routes

1. Create a new file in `app/routes/`
2. Add the route to `app/routes.ts`
3. Implement the component with proper TypeScript types

### Styling Best Practices

- Use Tailwind utilities for styling
- Follow responsive-first approach
- Maintain consistent spacing and colors
- Use semantic HTML elements

### Type Safety

- Define interfaces for all API responses
- Use Route types for loaders and actions
- Avoid `any` types where possible

## Performance Optimizations

- **Server-Side Rendering**: Fast initial page loads
- **Code Splitting**: Automatic route-based splitting
- **Optimized Images**: Responsive image handling
- **Efficient Bundling**: Vite-powered build process

## Browser Support

- Chrome/Edge 88+
- Firefox 85+
- Safari 14+
- Mobile browsers with ES2020 support

## Troubleshooting

### Common Issues

1. **Backend Connection Errors**
   - Ensure backend is running on port 8080
   - Check CORS configuration
   - Verify API endpoint URLs

2. **Build Errors**
   - Clear node_modules and reinstall
   - Check TypeScript errors with `npm run typecheck`
   - Verify all imports are correct

3. **Styling Issues**
   - Ensure Tailwind CSS is properly configured
   - Check for conflicting CSS classes
   - Verify responsive breakpoints

### Development Tips

- Use browser dev tools for debugging
- Check network tab for API call issues
- Use React Developer Tools for component debugging
- Monitor console for TypeScript errors

## Contributing

1. Follow TypeScript best practices
2. Use consistent naming conventions
3. Add proper error handling
4. Test responsive design
5. Document complex logic

## Deployment

### Development Deployment
```bash
npm run build
npm run start
```

### Production Deployment
The built application can be deployed to any Node.js hosting service:
- Vercel
- Netlify
- Railway
- Heroku
- Custom VPS

Make sure to:
1. Set proper environment variables
2. Configure backend API URLs
3. Enable HTTPS in production
4. Set up proper error monitoring

## Future Enhancements

- Real-time stock price updates
- Advanced charting capabilities
- Portfolio management features
- User authentication
- Saved screening strategies
- Export functionality
- Mobile app version