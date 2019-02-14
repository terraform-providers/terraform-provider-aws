exports.handler = async (event) => {
    if (process.env.TEST_DATA) {
        event.key3 = process.env.TEST_DATA;
    }
    return event;
}
