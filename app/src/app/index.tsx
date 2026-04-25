import { StyleSheet, Text, View } from 'react-native';
import { router } from 'expo-router';
import { Button } from '@/components/ui/Button';

export default function HomeScreen() {
    return (
        <View style={styles.container}>
            <Text style={styles.title}>Hello, Enterprise Architecture! 🚀</Text>
            <Text style={styles.subtitle}>Welcome to the new standard.</Text>

            <View style={styles.buttonContainer}>
                <Button
                    title="Go to Login"
                    onPress={() => router.push('/login')}
                />
                <Button
                    title="Learn More"
                    variant="secondary"
                    onPress={() => console.log('Pressed Secondary!')}
                />
            </View>
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
        fontSize: 24,
        fontWeight: 'bold',
        color: '#ffffff',
        marginBottom: 8,
        textAlign: 'center',
    },
    subtitle: {
        fontSize: 16,
        color: '#a1a1aa',
        marginBottom: 32,
    },
    buttonContainer: {
        width: '100%',
        maxWidth: 300,
        gap: 12, // React Native supports gap!
    }
});
