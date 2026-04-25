import React from 'react';
import {
    TouchableOpacity,
    Text,
    StyleSheet,
    ActivityIndicator,
    TouchableOpacityProps
} from 'react-native';

interface ButtonProps extends TouchableOpacityProps {
    title: string;
    variant?: 'primary' | 'secondary' | 'outline';
    isLoading?: boolean;
}

export function Button({
    title,
    variant = 'primary',
    isLoading = false,
    style,
    ...rest
}: ButtonProps) {
    const isPrimary = variant === 'primary';
    const isSecondary = variant === 'secondary';
    const isOutline = variant === 'outline';

    return (
        <TouchableOpacity
            style={[
                styles.button,
                isPrimary && styles.primary,
                isSecondary && styles.secondary,
                isOutline && styles.outline,
                style,
            ]}
            disabled={isLoading || rest.disabled}
            {...rest}
        >
            {isLoading ? (
                <ActivityIndicator color={isPrimary ? '#fff' : '#000'} />
            ) : (
                <Text style={[
                    styles.text,
                    isPrimary && styles.textPrimary,
                    isSecondary && styles.textSecondary,
                    isOutline && styles.textOutline,
                ]}>
                    {title}
                </Text>
            )}
        </TouchableOpacity>
    );
}

const styles = StyleSheet.create({
    button: {
        height: 48,
        borderRadius: 8,
        alignItems: 'center',
        justifyContent: 'center',
        paddingHorizontal: 16,
        flexDirection: 'row',
    },
    text: {
        fontSize: 16,
        fontWeight: '600',
    },
    // Primary Variant
    primary: {
        backgroundColor: '#ffffff', // White button on dark background
    },
    textPrimary: {
        color: '#000000',
    },
    // Secondary Variant
    secondary: {
        backgroundColor: '#333333',
    },
    textSecondary: {
        color: '#ffffff',
    },
    // Outline Variant
    outline: {
        backgroundColor: 'transparent',
        borderWidth: 1,
        borderColor: '#333333',
    },
    textOutline: {
        color: '#ffffff',
    },
});
