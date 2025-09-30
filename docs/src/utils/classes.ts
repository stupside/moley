// Component utility functions
export function cn(...classes: (string | undefined)[]): string {
    return classes.filter(Boolean).join(' ');
}

export function mergeClassName(baseClasses: string, className?: string): string {
    return cn(baseClasses, className);
}