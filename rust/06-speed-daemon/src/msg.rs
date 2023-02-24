#[derive(Clone, Debug, PartialEq, Eq, serde::Deserialize, serde::Serialize)]
pub struct Ticket {
    pub plate: String,
    pub road: u16,
    pub p1: Point,
    pub p2: Point,
    pub speed: u16,
}

#[derive(Clone, Debug, PartialEq, Eq, serde::Deserialize, serde::Serialize)]
pub struct Point(u32, u32);

#[derive(Clone, Debug, PartialEq, Eq, serde::Deserialize, serde::Serialize)]
pub struct IAmCamera {
    pub road: u16,
    pub mile: u16,
    pub limit: u16,
}

#[derive(Clone, Debug, PartialEq, Eq, serde::Deserialize, serde::Serialize)]
pub struct IAmDispatcher {
    pub roads: Vec<u16>,
}

#[derive(Clone, Debug, PartialEq, Eq, serde::Deserialize, serde::Serialize)]
pub struct WantHeartbeat {
    pub interval: u32,
}

#[derive(Clone, Debug, PartialEq, Eq, serde::Deserialize, serde::Serialize)]
pub struct Plate {
    pub plate: String,
    pub timestamp: u32,
}
