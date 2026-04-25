import { StyleSheet, View, Text } from 'react-native';
import { LoginForm } from '@/features/authentication/components/LoginForm';
import { useAuthStore } from '@/features/authentication/store/useAuthStore';
import { Button } from '@/components/ui/Button';
import { router } from 'expo-router';

export default function LoginScreen() {
    const { user, isAuthenticated, logout } = useAuthStore();

    return (
        <View style={styles.container}>
            <Text style={styles.title}>
                {isAuthenticated ? `Welcome back, ${user?.name}!` : 'Authentication'}
            </Text>

            {isAuthenticated ? (
                <View style={styles.loggedInContainer}>
                    <Button title="Go Home" variant="outline" onPress={() => router.push('/')} style={{ marginBottom: 12 }} />
                    <Button title="Log Out" variant="secondary" onPress={logout} />
                </View>
            ) : (
                <LoginForm />
            )}
        </View>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        backgroundColor: '#0a0a0a',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 24,
    },
    title: {
        fontSize: 28,
        fontWeight: 'bold',
        color: '#ffffff',
        marginBottom: 32,
    },
    loggedInContainer: {
        width: '100%',
        maxWidth: 300,
    }
});
