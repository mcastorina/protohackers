#[derive(Debug, serde::Deserialize)]
pub struct PrimeTimeInput {
    method: String,
    number: f64,
}

pub struct PrimeTime {
    number: Number,
}

#[derive(Debug, serde::Serialize)]
pub struct PrimeTimeOutput {
    method: &'static str,
    number: Number,
    prime: bool,
}

#[derive(Debug, serde::Serialize)]
#[serde(untagged)]
pub enum Number {
    Int(u64),
    Float(f64),
}

impl Number {
    pub fn as_int(&self) -> Option<u64> {
        match self {
            Self::Int(n) => Some(*n),
            _ => None,
        }
    }
}

#[derive(Debug, serde::Serialize)]
#[serde(tag = "error", rename_all = "snake_case")]
pub enum PrimeTimeError {
    InvalidInput,
    UnexpectedMethod,
}

impl<E> TryFrom<Result<PrimeTimeInput, E>> for PrimeTime {
    type Error = PrimeTimeError;

    fn try_from(input: Result<PrimeTimeInput, E>) -> Result<Self, Self::Error> {
        let input = input.map_err(|_| PrimeTimeError::InvalidInput)?;
        if input.method != "isPrime" {
            return Err(PrimeTimeError::UnexpectedMethod);
        }
        Ok(if input.number.fract() != 0. {
            PrimeTime {
                number: Number::Float(input.number),
            }
        } else {
            PrimeTime {
                number: Number::Int(input.number as u64),
            }
        })
    }
}

impl From<PrimeTime> for PrimeTimeOutput {
    fn from(prime: PrimeTime) -> Self {
        PrimeTimeOutput {
            method: "isPrime",
            prime: prime.number.as_int().map(primes::is_prime).unwrap_or(false),
            number: prime.number,
        }
    }
}
