
// read arg from command line 
const envFilter = process.argv[2];

// read all environment variables and print them

let envs = [];
for (const [key, value] of Object.entries(process.env)) {

    if (envFilter && !key.toLowerCase().includes(envFilter.toLowerCase())) {
        continue;
    }
    const vl = value.length > 40 ? value.substring(0, 40) + '...' : value;
    envs.push({ "Environment Variable": key, Value: vl });
}
console.clear();

if (envFilter) {
    console.log(`Filtering environment variables for: ${envFilter}`);
}

console.table(envs.reverse());