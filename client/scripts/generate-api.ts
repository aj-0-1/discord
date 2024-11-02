const { generate } = require('openapi-typescript-codegen');

(async () => {
    try {
        await generate({
            input: 'http://localhost:8080/swagger/doc.json',
            output: './src/api',
            httpClient: 'axios',
            clientName: 'ApiClient'
        });
        console.log('API types generated successfully');
    } catch (error) {
        console.error('Error generating API types:', error);
        process.exit(1);
    }
})();
