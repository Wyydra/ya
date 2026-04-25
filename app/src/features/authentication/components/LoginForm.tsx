import React, { useState } from 'react';
import { View, Text, TextInput, StyleSheet } from 'react-native';
import { Button } from '@/components/ui/Button';
import { useAuthStore } from '../store/useAuthStore';

export function LoginForm() {
    const [email, setEmail] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const login = useAuthStore((state) => state.login);

    const handleLogin = async () => {
        if (!email) return;
        setIsLoading(true);
        await login(email);
        setIsLoading(false);
    };

    return (
        <View style={styles.container}>
            <Text style={styles.label}>Email Address</Text>
            <TextInput
                style={styles.input}
                placeholder="Enter your email"
                placeholderTextColor="#666"
                value={email}
                onChangeText={setEmail}
                autoCapitalize="none"
                keyboardType="email-address"
            />
            <Button
                title="Sign In"
                onPress={handleLogin}
                isLoading={isLoading}
            />
        </View>
    );
}

const styles = StyleSheet.create({
    container: {
        width: '100%',
        padding: 24,
        backgroundColor: '#1a1a1a',
        borderRadius: 12,
        borderWidth: 1,
        borderColor: '#333',
    },
    label: {
        color: '#fff',
        marginBottom: 8,
        fontWeight: '500',
    },
    input: {
        backgroundColor: '#000',
        borderWidth: 1,
        borderColor: '#333',
        borderRadius: 8,
        padding: 16,
        color: '#fff',
        marginBottom: 16,
        fontSize: 16,
    },
});
