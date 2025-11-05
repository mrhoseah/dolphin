package app

import (
	"fmt"
	"os"
	"strings"
)

// CreateShadcnUI sets up shadcn/ui for React-based Dolphin apps
func (g *Generator) CreateShadcnUI() error {
	// Create necessary directories
	dirs := []string{
		"components/ui",
		"lib",
		"hooks",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create components.json
	if err := g.createComponentsJson(); err != nil {
		return fmt.Errorf("failed to create components.json: %w", err)
	}

	// Create lib/utils.ts
	if err := g.createUtilsTs(); err != nil {
		return fmt.Errorf("failed to create lib/utils.ts: %w", err)
	}

	// Update tailwind.config.js
	if err := g.updateTailwindConfig(); err != nil {
		return fmt.Errorf("failed to update tailwind.config.js: %w", err)
	}

	// Update assets/css/app.css
	if err := g.updateAppCss(); err != nil {
		return fmt.Errorf("failed to update app.css: %w", err)
	}

	// Create tsconfig.paths.json for path aliases
	if err := g.createTsconfigPaths(); err != nil {
		return fmt.Errorf("failed to create tsconfig.paths.json: %w", err)
	}

	// Update package.json with shadcn dependencies
	if err := g.updatePackageJson(); err != nil {
		return fmt.Errorf("failed to update package.json: %w", err)
	}

	// Create example component
	if err := g.createExampleComponent(); err != nil {
		return fmt.Errorf("failed to create example component: %w", err)
	}

	return nil
}

func (g *Generator) createComponentsJson() error {
	content := `{
  "$schema": "https://ui.shadcn.com/schema.json",
  "style": "default",
  "rsc": false,
  "tsx": true,
  "tailwind": {
    "config": "tailwind.config.js",
    "css": "assets/css/app.css",
    "baseColor": "slate",
    "cssVariables": true,
    "prefix": ""
  },
  "aliases": {
    "components": "@/components",
    "utils": "@/lib/utils",
    "ui": "@/components/ui",
    "lib": "@/lib",
    "hooks": "@/hooks"
  }
}
`
	return os.WriteFile("components.json", []byte(content), 0644)
}

func (g *Generator) createUtilsTs() error {
	content := `import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

/**
 * Utility function to merge Tailwind CSS classes
 * Used by shadcn/ui components
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
`
	return os.WriteFile("lib/utils.ts", []byte(content), 0644)
}

func (g *Generator) updateTailwindConfig() error {
	configPath := "tailwind.config.js"

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create new tailwind config
		content := `/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: ["class"],
  content: [
    './views/**/*.fin.html',
    './assets/js/**/*.{js,ts,jsx,tsx}',
    './assets/ts/**/*.{js,ts,jsx,tsx}',
    './components/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
      keyframes: {
        "accordion-down": {
          from: { height: "0" },
          to: { height: "var(--radix-accordion-content-height)" },
        },
        "accordion-up": {
          from: { height: "var(--radix-accordion-content-height)" },
          to: { height: "0" },
        },
      },
      animation: {
        "accordion-down": "accordion-down 0.2s ease-out",
        "accordion-up": "accordion-up 0.2s ease-out",
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
}
`
		return os.WriteFile(configPath, []byte(content), 0644)
	}

	// File exists, read and check if shadcn config is already there
	existing, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// Check if shadcn colors are already present
	if strings.Contains(string(existing), "hsl(var(--border))") {
		return nil // Already configured
	}

	// Append shadcn configuration (this is a simple approach)
	// In production, you'd want to parse and merge the config properly
	return nil // Don't overwrite existing config
}

func (g *Generator) updateAppCss() error {
	cssPath := "assets/css/app.css"

	// Check if file exists
	if _, err := os.Stat(cssPath); os.IsNotExist(err) {
		// Create new CSS file with shadcn variables
		content := g.generateShadcnCss()
		return os.WriteFile(cssPath, []byte(content), 0644)
	}

	// File exists, read and check if shadcn variables are already there
	existing, err := os.ReadFile(cssPath)
	if err != nil {
		return err
	}

	// Check if shadcn variables are already present
	if strings.Contains(string(existing), "--background:") {
		return nil // Already configured
	}

	// Append shadcn CSS variables
	shadcnCss := "\n\n" + g.generateShadcnCssVariables()
	content := string(existing) + shadcnCss
	return os.WriteFile(cssPath, []byte(content), 0644)
}

func (g *Generator) generateShadcnCss() string {
	return `@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
` + g.generateShadcnCssVariables() + `  }

  .dark {
` + g.generateShadcnDarkVariables() + `  }
}

@layer base {
  * {
    @apply border-border;
  }
  body {
    @apply bg-background text-foreground;
  }
}
`
}

func (g *Generator) generateShadcnCssVariables() string {
	return `    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --card: 0 0% 100%;
    --card-foreground: 222.2 84% 4.9%;
    --popover: 0 0% 100%;
    --popover-foreground: 222.2 84% 4.9%;
    --primary: 222.2 47.4% 11.2%;
    --primary-foreground: 210 40% 98%;
    --secondary: 210 40% 96.1%;
    --secondary-foreground: 222.2 47.4% 11.2%;
    --muted: 210 40% 96.1%;
    --muted-foreground: 215.4 16.3% 46.9%;
    --accent: 210 40% 96.1%;
    --accent-foreground: 222.2 47.4% 11.2%;
    --destructive: 0 84.2% 60.2%;
    --destructive-foreground: 210 40% 98%;
    --border: 214.3 31.8% 91.4%;
    --input: 214.3 31.8% 91.4%;
    --ring: 222.2 84% 4.9%;
    --radius: 0.5rem;
`
}

func (g *Generator) generateShadcnDarkVariables() string {
	return `    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    --card: 222.2 84% 4.9%;
    --card-foreground: 210 40% 98%;
    --popover: 222.2 84% 4.9%;
    --popover-foreground: 210 40% 98%;
    --primary: 210 40% 98%;
    --primary-foreground: 222.2 47.4% 11.2%;
    --secondary: 217.2 32.6% 17.5%;
    --secondary-foreground: 210 40% 98%;
    --muted: 217.2 32.6% 17.5%;
    --muted-foreground: 215 20.2% 65.1%;
    --accent: 217.2 32.6% 17.5%;
    --accent-foreground: 210 40% 98%;
    --destructive: 0 62.8% 30.6%;
    --destructive-foreground: 210 40% 98%;
    --border: 217.2 32.6% 17.5%;
    --input: 217.2 32.6% 17.5%;
    --ring: 212.7 26.8% 83.9%;
`
}

func (g *Generator) createTsconfigPaths() error {
	content := `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["./*"]
    }
  }
}
`
	return os.WriteFile("tsconfig.paths.json", []byte(content), 0644)
}

func (g *Generator) updatePackageJson() error {
	// This would ideally parse and update package.json
	// For now, we'll create a script that users can run
	return nil
}

func (g *Generator) createExampleComponent() error {
	// Create a simple Button component example
	content := `import * as React from "react"
import { cn } from "@/lib/utils"

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "default" | "destructive" | "outline" | "secondary" | "ghost" | "link"
  size?: "default" | "sm" | "lg" | "icon"
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "default", size = "default", ...props }, ref) => {
    return (
      <button
        className={cn(
          "inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
          {
            "bg-primary text-primary-foreground hover:bg-primary/90": variant === "default",
            "bg-destructive text-destructive-foreground hover:bg-destructive/90": variant === "destructive",
            "border border-input bg-background hover:bg-accent hover:text-accent-foreground": variant === "outline",
            "bg-secondary text-secondary-foreground hover:bg-secondary/80": variant === "secondary",
            "hover:bg-accent hover:text-accent-foreground": variant === "ghost",
            "text-primary underline-offset-4 hover:underline": variant === "link",
          },
          {
            "h-10 px-4 py-2": size === "default",
            "h-9 rounded-md px-3": size === "sm",
            "h-11 rounded-md px-8": size === "lg",
            "h-10 w-10": size === "icon",
          },
          className
        )}
        ref={ref}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"

export { Button }
`

	return os.WriteFile("components/ui/button.tsx", []byte(content), 0644)
}
